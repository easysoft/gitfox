# GitFox Open Source

> Based on the secondary development of the open-source version of [Gitness](https://github.com/harness/harness)

## Overview
Harness Open source is an open source development platform packed with the power of code hosting, automated DevOps pipelines.


## Running GitFox locally

To install GitFox yourself, simply run the command below. Once the container is up, you can visit http://localhost:3000 in your browser.

```bash
docker run -d \
  -p 3000:3000 \
  -p 3022:3022 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /tmp/gitfox:/data \
  --name gitfox \
  --restart always \
  hub.zentao.net/app/gitfox-oss:oss
```
> The GitFox image uses a volume to store the database and repositories. It is highly recommended to use a bind mount or named volume as otherwise all data will be lost once the container is stopped.
