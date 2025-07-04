module github.com/easysoft/gitfox

go 1.22.0

replace (
	github.com/drone/drone-go => github.com/quicklyon/drone-go v1.7.3
	github.com/drone/runner-go => github.com/quicklyon/runner-go v1.12.6
)

require (
	cloud.google.com/go/storage v1.42.0
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/semver/v3 v3.2.1
	github.com/Masterminds/squirrel v1.5.4
	github.com/adrg/xdg v0.5.0
	github.com/antonmedv/expr v1.15.5
	github.com/aws/aws-sdk-go v1.55.2
	github.com/bmatcuk/doublestar v1.3.4
	github.com/bmatcuk/doublestar/v4 v4.6.1
	github.com/buildkite/yaml v2.1.0+incompatible
	github.com/dchest/uniuri v1.2.0
	github.com/docker/docker v24.0.7+incompatible
	github.com/docker/go-connections v0.5.0
	github.com/docker/go-units v0.5.0
	github.com/drone/drone-go v1.7.1
	github.com/drone/funcmap v0.0.0-20190918184546-d4ef6e88376d
	github.com/drone/go-convert v0.0.0-20240821195621-c6d7be7727ec
	github.com/drone/go-generate v0.0.0-20230920014042-6085ee5c9522
	github.com/drone/go-scm v1.38.4
	github.com/drone/runner-go v1.12.6
	github.com/fatih/color v1.17.0
	github.com/gabriel-vasile/mimetype v1.4.4
	github.com/ghodss/yaml v1.0.0
	github.com/gliderlabs/ssh v0.3.7
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-redsync/redsync/v4 v4.13.0
	github.com/go-sql-driver/mysql v1.8.1
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/golang-module/carbon/v2 v2.3.10
	github.com/google/go-cmp v0.6.0
	github.com/google/go-jsonnet v0.20.0
	github.com/google/uuid v1.6.0
	github.com/google/wire v0.6.0
	github.com/gorhill/cronexpr v0.0.0-20180427100037-88b0669f7d75
	github.com/gotidy/ptr v1.4.0
	github.com/guregu/null v4.0.0+incompatible
	github.com/harness/harness-migrate v0.26.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-version v1.6.0
	github.com/jackc/pgerrcode v0.0.0-20240316143900-6e2875d9b438
	github.com/jackc/pgx/v5 v5.5.5 // indirect
	github.com/jmoiron/sqlx v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lib/pq v1.10.9
	github.com/mailru/easyjson v0.7.7
	github.com/matoous/go-nanoid v1.5.0
	github.com/matoous/go-nanoid/v2 v2.1.0
	github.com/mattn/go-isatty v0.0.20
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/natessilva/dag v0.0.0-20180124060714-7194b8dcc5c4
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/quicklyon/kingpin/v2 v2.2.7-0.20240528063919-2c35cbd0f246
	github.com/rs/xid v1.5.0
	github.com/rs/zerolog v1.33.0
	github.com/sashabaranov/go-openai v1.36.0
	github.com/satori/go.uuid v1.2.0
	github.com/sercand/kuberesolver/v5 v5.1.1
	github.com/sirupsen/logrus v1.9.3
	github.com/slack-go/slack v0.14.0
	github.com/stretchr/testify v1.9.0
	github.com/swaggest/openapi-go v0.2.23
	github.com/swaggest/swgui v1.8.1
	github.com/unrolled/secure v1.15.0
	github.com/zricethezav/gitleaks/v8 v8.18.5-0.20240912004812-e93a7c0d2604
	go.starlark.net v0.0.0-20231121155337-90ade8b19d09
	go.uber.org/multierr v1.11.0
	golang.org/x/crypto v0.27.0
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56
	golang.org/x/oauth2 v0.21.0
	golang.org/x/sync v0.8.0
	golang.org/x/term v0.24.0
	golang.org/x/text v0.18.0
	google.golang.org/api v0.189.0
	gopkg.in/DATA-DOG/go-sqlmock.v2 v2.0.0-20180914054222-c19298f520d0
	gopkg.in/mail.v2 v2.3.1
	gorm.io/driver/mysql v1.5.7
	gorm.io/driver/postgres v1.5.4
	gorm.io/driver/sqlite v1.5.4
	gorm.io/gorm v1.25.12
	helm.sh/helm/v3 v3.10.0
	k8s.io/api v0.28.2 // indirect
	k8s.io/apimachinery v0.28.2 // indirect
	k8s.io/client-go v0.28.2
	sigs.k8s.io/yaml v1.4.0
)

require (
	cloud.google.com/go v0.115.0 // indirect
	cloud.google.com/go/auth v0.7.2 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.3 // indirect
	cloud.google.com/go/compute/metadata v0.5.0 // indirect
	cloud.google.com/go/iam v1.1.12 // indirect
	dario.cat/mergo v1.0.1 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/99designs/httpsignatures-go v0.0.0-20170731043157-88528bf4ca7e // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/BobuSumisu/aho-corasick v1.0.3 // indirect
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/charmbracelet/lipgloss v0.12.1 // indirect
	github.com/charmbracelet/x/ansi v0.1.4 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/drone/envsubst v1.0.3 // indirect
	github.com/drone/spec v0.0.0-20230920145636-3827abdce961 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/fatih/semgroup v1.2.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gitleaks/go-gitdiff v0.9.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.20.2 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.8 // indirect
	github.com/go-test/deep v1.0.8 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/s2a-go v0.1.8 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.13.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/h2non/filetype v1.1.3 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/muesli/termenv v0.15.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/onsi/ginkgo/v2 v2.11.0 // indirect
	github.com/onsi/gomega v1.27.10 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/prometheus/client_golang v1.19.1 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.55.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/sagikazarmark/locafero v0.6.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.19.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.49.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.51.0 // indirect
	go.opentelemetry.io/otel v1.30.0 // indirect
	go.opentelemetry.io/otel/metric v1.30.0 // indirect
	go.opentelemetry.io/otel/sdk v1.30.0 // indirect
	go.opentelemetry.io/otel/trace v1.30.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/grpc v1.66.1 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gotest.tools/v3 v3.5.1 // indirect
	k8s.io/klog/v2 v2.100.1 // indirect
	k8s.io/kube-openapi v0.0.0-20230717233707-2695361300d9 // indirect
	k8s.io/utils v0.0.0-20230406110748-d93618cff8a2 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)

require (
	github.com/docker/distribution v2.8.1+incompatible
	github.com/go-chi/chi/v5 v5.2.0
	github.com/go-chi/cors v1.2.1
	github.com/go-logr/logr v1.4.2
	github.com/lucasb-eyer/go-colorful v1.2.0
)

require (
	github.com/go-logr/zerologr v1.2.3
	github.com/mattn/go-colorable v0.1.13 // indirect
)

require (
	cloud.google.com/go/profiler v0.3.1
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20231202071711-9a357b53e9c9 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/djherbis/buffer v1.2.0
	github.com/djherbis/nio/v3 v3.0.1
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/pprof v0.0.0-20240722153945-304e4f0156b8 // indirect
	github.com/google/subcommands v1.2.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/swaggest/jsonschema-go v0.3.40
	github.com/swaggest/refl v1.1.0 // indirect
	github.com/vearutop/statigz v1.4.0 // indirect
	github.com/yuin/goldmark v1.4.13
	golang.org/x/mod v0.19.0 // indirect
	golang.org/x/net v0.29.0
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/tools v0.23.0 // indirect
	google.golang.org/genproto v0.0.0-20240722135656-d784300faade // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1
)
