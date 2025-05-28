# Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
# Use of this source code is covered by the following dual licenses:
# (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
# (2) Affero General Public License 3.0 (AGPL 3.0)
# license that can be found in the LICENSE file.

FROM hub.zentao.net/ci/node:20.11.1-bookworm-slim AS web

ARG TZ="Asia/Shanghai"

ENV TZ ${TZ}

WORKDIR /usr/src/app

COPY web/package.json ./

COPY web/yarn.lock ./

COPY ./web .

RUN yarn config set registry https://mirrors.huaweicloud.com/repository/npm/ && yarn && yarn build && yarn cache clean

FROM hub.zentao.net/ci/golang:1.22.1-alpine AS build

ARG TZ="Asia/Shanghai"

ENV TZ ${TZ}

ENV GOPROXY=https://goproxy.cn,direct

ENV GO111MODULE=on

WORKDIR /app

COPY . .

COPY --from=web /usr/src/app/dist /app/web/dist

RUN go install github.com/go-task/task/v3/cmd/task@latest \
  && task build \
  && task download-gitfox-plugin

### Pull CA Certs
FROM hub.zentao.net/ci/alpine:3.19 AS cert-image

RUN apk --update add ca-certificates

FROM hub.zentao.net/ci/alpine:3.19

ARG TZ="Asia/Shanghai"

ENV TZ ${TZ}

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories \
    && apk --no-cache add git bash musl tzdata curl \
    && ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone \
    && date \
    # && apk del --no-cache tzdata \
    && rm -rf /var/cache/apk/*

# setup app dir and its content
WORKDIR /app

VOLUME /data

ENV XDG_CACHE_HOME /data
ENV GITFOX_STORAGE_DIR /data
ENV GITFOX_GIT_ROOT /data
# ENV GITFOX_CI_PLUGINS_ZIP_URL=https://pkg.zentao.net/gitfox/20241024/plugins.zip
ENV GITFOX_CI_PLUGINS_ZIP_URL=/app/plugins.zip
ENV GITFOX_GIT_DEFAULTBRANCH=master
ENV GITFOX_DEBUG=true
ENV GITFOX_TRACE=true
ENV GITFOX_GIT_TRACE=true
ENV GITFOX_WEBHOOK_ALLOW_LOOPBACK=true
ENV GITFOX_WEBHOOK_ALLOW_PRIVATE_NETWORK=true
ENV GITFOX_METRIC_ENABLED=false
ENV GITFOX_METRIC_ENDPOINT=https://stats.drone.ci/api/v1/gitness
ENV GITFOX_TOKEN_COOKIE_NAME=token
ENV GITFOX_DOCKER_HOST unix:///var/run/docker.sock
ENV GITFOX_DOCKER_API_VERSION 1.40
ENV GITFOX_SSH_ENABLE=true
ENV GITFOX_SSH_HOST_KEYS_DIR=/data/ssh
ENV GITFOX_GITSPACE_ENABLE=false

COPY --from=build /app/bin/gitfox_linux_amd64 /app/gitfox

COPY --from=build /app/hack/ci-plugin/plugins.zip /app/plugins.zip

COPY --from=cert-image /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE 3000
EXPOSE 22

CMD [ "/app/gitfox", "server" ]
