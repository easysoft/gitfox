// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"time"

	"github.com/easysoft/gitfox/blob"
	"github.com/easysoft/gitfox/events"
	gitenum "github.com/easysoft/gitfox/git/enum"
	"github.com/easysoft/gitfox/lock"
	"github.com/easysoft/gitfox/pkg/storage"
	"github.com/easysoft/gitfox/pubsub"

	gossh "golang.org/x/crypto/ssh"
)

type StorageProviderType string

const (
	StorageProviderLocal StorageProviderType = "local"
	StorageProviderS3    StorageProviderType = "s3"
)

// Config stores the system configuration.
type Config struct {
	// InstanceID specifis the ID of the Gitfox instance.
	// NOTE: If the value is not provided the hostname of the machine is used.
	InstanceID string `envconfig:"GITFOX_INSTANCE_ID"`

	Debug bool `envconfig:"GITFOX_DEBUG"`
	Trace bool `envconfig:"GITFOX_TRACE"`

	// GracefulShutdownTime defines the max time we wait when shutting down a server.
	// 5min should be enough for most git clones to complete.
	GracefulShutdownTime time.Duration `envconfig:"GITFOX_GRACEFUL_SHUTDOWN_TIME" default:"300s"`

	UserSignupEnabled   bool `envconfig:"GITFOX_USER_SIGNUP_ENABLED" default:"true"`
	NestedSpacesEnabled bool `envconfig:"GITFOX_NESTED_SPACES_ENABLED" default:"false"`

	// PublicResourceCreationEnabled specifies whether a user can create publicly accessible resources.
	PublicResourceCreationEnabled bool `envconfig:"GITFOX_PUBLIC_RESOURCE_CREATION_ENABLED" default:"true"`

	Profiler struct {
		Type        string `envconfig:"GITFOX_PROFILER_TYPE"`
		ServiceName string `envconfig:"GITFOX_PROFILER_SERVICE_NAME" default:"gitfox"`
	}

	// URL defines the URLs via which the different parts of the service are reachable by.
	URL struct {
		// Base is used to generate external facing URLs in case they aren't provided explicitly.
		// Value is derived from Server.HTTP Config unless explicitly specified (e.g. http://localhost:3000).
		Base string `envconfig:"GITFOX_URL_BASE"`

		// Git defines the external URL via which the GIT API is reachable.
		// NOTE: for routing to work properly, the request path & hostname reaching gitfox
		// have to statisfy at least one of the following two conditions:
		// - Path ends with `/git`
		// - Hostname is different to API hostname
		// (this could be after proxy path / header rewrite).
		// Value is derived from Base unless explicitly specified (e.g. http://localhost:3000/git).
		Git string `envconfig:"GITFOX_URL_GIT"`

		// GitSSH defines the external URL via which the GIT SSH server is reachable.
		// Value is derived from Base or SSH Config unless explicitly specified (e.g. ssh://localhost).
		GitSSH string `envconfig:"GITFOX_URL_GIT_SSH"`

		// API defines the external URL via which the rest API is reachable.
		// NOTE: for routing to work properly, the request path reaching gitfox has to end with `/api`
		// (this could be after proxy path rewrite).
		// Value is derived from Base unless explicitly specified (e.g. http://localhost:3000/api).
		API string `envconfig:"GITFOX_URL_API"`

		// UI defines the external URL via which the UI is reachable.
		// Value is derived from Base unless explicitly specified (e.g. http://localhost:3000).
		UI string `envconfig:"GITFOX_URL_UI"`

		// Internal defines the internal URL via which the service is reachable.
		// Value is derived from HTTP.Server unless explicitly specified (e.g. http://localhost:3000).
		Internal string `envconfig:"GITFOX_URL_INTERNAL"`

		// Container is the endpoint that can be used by running container builds to communicate
		// with Gitfox (for example while performing a clone on a local repo).
		// host.docker.internal allows a running container to talk to services exposed on the host
		// (either running directly or via a port exposed in a docker container).
		// Value is derived from HTTP.Server unless explicitly specified (e.g. http://host.docker.internal:3000).
		Container string `envconfig:"GITFOX_URL_CONTAINER"`

		// Registry is used as a base to generate external facing URLs.
		// Value is derived from HTTP.Server unless explicitly specified (e.g. http://host.docker.internal:3000).
		Registry string `envconfig:"GITFOX_URL_REGISTRY"`
	}

	// Git defines the git configuration parameters
	Git struct {
		// Trace specifies whether git operations should be traces.
		// NOTE: Currently limited to 'push' operation until we move to internal command package.
		Trace bool `envconfig:"GITFOX_GIT_TRACE"`
		// DefaultBranch specifies the default branch for new repositories.
		DefaultBranch string `envconfig:"GITFOX_GIT_DEFAULTBRANCH" default:"main"`
		// Root specifies the directory containing git related data (e.g. repos, ...)
		Root string `envconfig:"GITFOX_GIT_ROOT"`
		// TmpDir (optional) specifies the directory for temporary data (e.g. repo clones, ...)
		TmpDir string `envconfig:"GITFOX_GIT_TMP_DIR"`
		// HookPath points to the binary used as git server hook.
		HookPath string `envconfig:"GITFOX_GIT_HOOK_PATH"`

		// LastCommitCache holds configuration options for the last commit cache.
		LastCommitCache struct {
			// Mode determines where the cache will be. Valid values are "inmemory" (default), "redis" or "none".
			Mode gitenum.LastCommitCacheMode `envconfig:"GITFOX_GIT_LAST_COMMIT_CACHE_MODE" default:"inmemory"`

			// Duration defines cache duration of last commit.
			Duration time.Duration `envconfig:"GITFOX_GIT_LAST_COMMIT_CACHE_DURATION" default:"12h"`
		}
	}

	// Encrypter defines the parameters for the encrypter
	Encrypter struct {
		Secret       string `envconfig:"GITFOX_ENCRYPTER_SECRET"` // key used for encryption
		MixedContent bool   `envconfig:"GITFOX_ENCRYPTER_MIXED_CONTENT"`
	}

	// HTTP defines the http server configuration parameters
	HTTP struct {
		Port  int    `envconfig:"GITFOX_HTTP_PORT" default:"3000"`
		Host  string `envconfig:"GITFOX_HTTP_HOST"`
		Proto string `envconfig:"GITFOX_HTTP_PROTO" default:"http"`
	}

	// Acme defines Acme configuration parameters.
	Acme struct {
		Enabled bool   `envconfig:"GITFOX_ACME_ENABLED"`
		Endpont string `envconfig:"GITFOX_ACME_ENDPOINT"`
		Email   bool   `envconfig:"GITFOX_ACME_EMAIL"`
		Host    string `envconfig:"GITFOX_ACME_HOST"`
	}

	SSH struct {
		Enable bool   `envconfig:"GITFOX_SSH_ENABLE" default:"false"`
		Host   string `envconfig:"GITFOX_SSH_HOST"`
		Port   int    `envconfig:"GITFOX_SSH_PORT" default:"3022"`
		// DefaultUser holds value for generating urls {user}@host:path and force check
		// no other user can authenticate unless it is empty then any username is allowed
		DefaultUser             string   `envconfig:"GITFOX_SSH_DEFAULT_USER" default:"git"`
		Ciphers                 []string `envconfig:"GITFOX_SSH_CIPHERS"`
		KeyExchanges            []string `envconfig:"GITFOX_SSH_KEY_EXCHANGES"`
		MACs                    []string `envconfig:"GITFOX_SSH_MACS"`
		ServerHostKeys          []string `envconfig:"GITFOX_SSH_HOST_KEYS"`
		TrustedUserCAKeys       []string `envconfig:"GITFOX_SSH_TRUSTED_USER_CA_KEYS"`
		TrustedUserCAKeysFile   string   `envconfig:"GITFOX_SSH_TRUSTED_USER_CA_KEYS_FILENAME"`
		TrustedUserCAKeysParsed []gossh.PublicKey
		KeepAliveInterval       time.Duration `envconfig:"GITFOX_SSH_KEEP_ALIVE_INTERVAL" default:"5s"`
		ServerHostKeysDir       string        `envconfig:"GITFOX_SSH_HOST_KEYS_DIR"`
	}

	// CI defines configuration related to build executions.
	CI struct {
		ParallelWorkers int `envconfig:"GITFOX_CI_PARALLEL_WORKERS" default:"2"`
		// PluginsZipURL is a pointer to a zip containing all the plugins schemas.
		// This could be a local path or an external location.
		// If not provided, the default value is used. https://github.com/bradrydzewski/plugins/archive/refs/heads/master.zip
		//nolint:lll
		PluginsZipURL string `envconfig:"GITFOX_CI_PLUGINS_ZIP_URL" default:"https://pkg.zentao.net/gitfox/20241024/plugins.zip"`

		// ContainerNetworks is a list of networks that all containers created as part of CI
		// should be attached to.
		// This can be needed when we don't want to use host.docker.internal (eg when a service mesh
		// or proxy is being used) and instead want all the containers to run on the same network as
		// the gitfox container so that they can interact via the container name.
		// In that case, GITFOX_URL_CONTAINER should also be changed
		// (eg to http://<gitfox_container_name>:<port>).
		ContainerNetworks []string `envconfig:"GITFOX_CI_CONTAINER_NETWORKS"`

		Runner string `envconfig:"GITFOX_CI_RUNNER" default:"docker"`

		Kubernetes struct {
			Namespace      string `envconfig:"GITFOX_CI_KUBE_NAMESPACE" default:"quickon-ci"`
			ServiceAccount string `envconfig:"GITFOX_CI_KUBE_SERVICE_ACCOUNT" default:"default"`
		}

		Storage struct {
			Provider StorageProviderType `envconfig:"GITFOX_CI_STORAGE_PROVIDER" default:"local"`
			Prefix   string              `envconfig:"GITFOX_CI_STORAGE_PREFIX" default:"pipelines"`
		}
	}

	Artifact struct {
		Storage struct {
			Provider StorageProviderType `envconfig:"GITFOX_ARTIFACT_STORAGE_PROVIDER" default:"local"`
			Prefix   string              `envconfig:"GITFOX_ARTIFACT_STORAGE_PREFIX" default:"artifacts"`
		}
	}

	// Database defines the database configuration parameters.
	Database struct {
		Driver     string `envconfig:"GITFOX_DATABASE_DRIVER" default:"sqlite3"`
		Datasource string `envconfig:"GITFOX_DATABASE_DATASOURCE" default:"gitfox.db"`
		Host       string `envconfig:"GITFOX_DATABASE_HOST"`
		Port       int    `envconfig:"GITFOX_DATABASE_PORT"`
		User       string `envconfig:"GITFOX_DATABASE_USERNAME"`
		Password   string `envconfig:"GITFOX_DATABASE_PASSWORD"`
		DBName     string `envconfig:"GITFOX_DATABASE_DBNAME"`
		ExtraFlags string `envconfig:"GITFOX_DATABASE_EXTRA_OPTIONS"`
		Trace      bool   `envconfig:"GITFOX_DATABASE_TRACE" default:"false"`
	}

	// BlobStore defines the blob storage configuration parameters.
	BlobStore struct {
		// Provider is a name of blob storage service like filesystem or gcs
		Provider blob.Provider `envconfig:"GITFOX_BLOBSTORE_PROVIDER" default:"filesystem"`
		// Bucket is a path to the directory where the files will be stored when using filesystem blob storage,
		// in case of gcs provider this will be the actual bucket where the images are stored.
		Bucket string `envconfig:"GITFOX_BLOBSTORE_BUCKET"`

		// In case of GCS provider, this is expected to be the path to the service account key file.
		KeyPath string `envconfig:"GITFOX_BLOBSTORE_KEY_PATH" default:""`

		// Email ID of the google service account that needs to be impersonated
		TargetPrincipal string `envconfig:"GITFOX_BLOBSTORE_TARGET_PRINCIPAL" default:""`

		ImpersonationLifetime time.Duration `envconfig:"GITFOX_BLOBSTORE_IMPERSONATION_LIFETIME" default:"12h"`
	}

	Storage struct {
		Local struct {
			Directory string `envconfig:"GITFOX_STORAGE_DIR" default:"data"`
		}
		S3 struct {
			Driver storage.DriverType
			Host   string
			Region string
			Bucket string
		}
	}

	// Token defines token configuration parameters.
	Token struct {
		CookieName string        `envconfig:"GITFOX_TOKEN_COOKIE_NAME" default:"token"`
		Expire     time.Duration `envconfig:"GITFOX_TOKEN_EXPIRE" default:"720h"`
	}

	Logs struct {
		// S3 provides optional storage option for logs.
		S3 struct {
			Bucket    string `envconfig:"GITFOX_LOGS_S3_BUCKET"`
			Prefix    string `envconfig:"GITFOX_LOGS_S3_PREFIX"`
			Endpoint  string `envconfig:"GITFOX_LOGS_S3_ENDPOINT"`
			PathStyle bool   `envconfig:"GITFOX_LOGS_S3_PATH_STYLE"`
		}
	}

	// Cors defines http cors parameters
	Cors struct {
		AllowedOrigins   []string `envconfig:"GITFOX_CORS_ALLOWED_ORIGINS"   default:"*"`
		AllowedMethods   []string `envconfig:"GITFOX_CORS_ALLOWED_METHODS"   default:"GET,POST,PATCH,PUT,DELETE,OPTIONS"`
		AllowedHeaders   []string `envconfig:"GITFOX_CORS_ALLOWED_HEADERS"   default:"Origin,Accept,Accept-Language,Authorization,Content-Type,Content-Language,X-Requested-With,X-Request-Id"` //nolint:lll // struct tags can't be multiline
		ExposedHeaders   []string `envconfig:"GITFOX_CORS_EXPOSED_HEADERS"   default:"Link"`
		AllowCredentials bool     `envconfig:"GITFOX_CORS_ALLOW_CREDENTIALS" default:"true"`
		MaxAge           int      `envconfig:"GITFOX_CORS_MAX_AGE"           default:"300"`
	}

	// Secure defines http security parameters.
	Secure struct {
		AllowedHosts          []string          `envconfig:"GITFOX_HTTP_ALLOWED_HOSTS"`
		HostsProxyHeaders     []string          `envconfig:"GITFOX_HTTP_PROXY_HEADERS"`
		SSLRedirect           bool              `envconfig:"GITFOX_HTTP_SSL_REDIRECT"`
		SSLTemporaryRedirect  bool              `envconfig:"GITFOX_HTTP_SSL_TEMPORARY_REDIRECT"`
		SSLHost               string            `envconfig:"GITFOX_HTTP_SSL_HOST"`
		SSLProxyHeaders       map[string]string `envconfig:"GITFOX_HTTP_SSL_PROXY_HEADERS"`
		STSSeconds            int64             `envconfig:"GITFOX_HTTP_STS_SECONDS"`
		STSIncludeSubdomains  bool              `envconfig:"GITFOX_HTTP_STS_INCLUDE_SUBDOMAINS"`
		STSPreload            bool              `envconfig:"GITFOX_HTTP_STS_PRELOAD"`
		ForceSTSHeader        bool              `envconfig:"GITFOX_HTTP_STS_FORCE_HEADER"`
		BrowserXSSFilter      bool              `envconfig:"GITFOX_HTTP_BROWSER_XSS_FILTER"    default:"true"`
		FrameDeny             bool              `envconfig:"GITFOX_HTTP_FRAME_DENY"            default:"false"` // 默认true, 支持禅道嵌入改成false
		ContentTypeNosniff    bool              `envconfig:"GITFOX_HTTP_CONTENT_TYPE_NO_SNIFF"`
		ContentSecurityPolicy string            `envconfig:"GITFOX_HTTP_CONTENT_SECURITY_POLICY"`
		ReferrerPolicy        string            `envconfig:"GITFOX_HTTP_REFERRER_POLICY"`
	}

	Principal struct {
		// System defines the principal information used to create the system service.
		System struct {
			UID         string `envconfig:"GITFOX_PRINCIPAL_SYSTEM_UID"          default:"gitfox"`
			DisplayName string `envconfig:"GITFOX_PRINCIPAL_SYSTEM_DISPLAY_NAME" default:"Gitfox"`
			Email       string `envconfig:"GITFOX_PRINCIPAL_SYSTEM_EMAIL"        default:"system@gitfox.io"`
		}
		// Pipeline defines the principal information used to create the pipeline service.
		Pipeline struct {
			UID         string `envconfig:"GITFOX_PRINCIPAL_PIPELINE_UID"          default:"pipeline"`
			DisplayName string `envconfig:"GITFOX_PRINCIPAL_PIPELINE_DISPLAY_NAME" default:"Gitfox Pipeline"`
			Email       string `envconfig:"GITFOX_PRINCIPAL_PIPELINE_EMAIL"        default:"pipeline@gitfox.io"`
		}
		// Bot defines the principal information used to create the bot user.
		DefaultBot struct {
			UID         string `envconfig:"GITFOX_PRINCIPAL_BOT_UID"          default:"bot"`
			DisplayName string `envconfig:"GITFOX_PRINCIPAL_BOT_DISPLAY_NAME" default:"Gitfox Bot"`
			Email       string `envconfig:"GITFOX_PRINCIPAL_BOT_EMAIL"        default:"no_replay_bot@gitfox.io"`
		}

		// Gitspace defines the principal information used to create the gitspace service.
		Gitspace struct {
			UID         string `envconfig:"GITFOX_PRINCIPAL_GITSPACE_UID"          default:"gitspace"`
			DisplayName string `envconfig:"GITFOX_PRINCIPAL_GITSPACE_DISPLAY_NAME" default:"Gitfox Gitspace"`
			Email       string `envconfig:"GITFOX_PRINCIPAL_GITSPACE_EMAIL"        default:"gitspace@gitfox.io"`
		}

		// Admin defines the principal information used to create the admin user.
		// NOTE: The admin user is only auto-created in case a password and an email is provided.
		Admin struct {
			UID         string `envconfig:"GITFOX_PRINCIPAL_ADMIN_UID"           default:"admin"`
			DisplayName string `envconfig:"GITFOX_PRINCIPAL_ADMIN_DISPLAY_NAME"  default:"Administrator"`
			Email       string `envconfig:"GITFOX_PRINCIPAL_ADMIN_EMAIL"`    // No default email
			Password    string `envconfig:"GITFOX_PRINCIPAL_ADMIN_PASSWORD"` // No default password
		}
	}

	Redis struct {
		Endpoint           string `envconfig:"GITFOX_REDIS_ENDPOINT"              default:"localhost:6379"`
		MaxRetries         int    `envconfig:"GITFOX_REDIS_MAX_RETRIES"           default:"3"`
		MinIdleConnections int    `envconfig:"GITFOX_REDIS_MIN_IDLE_CONNECTIONS"  default:"0"`
		Password           string `envconfig:"GITFOX_REDIS_PASSWORD"`
		SentinelMode       bool   `envconfig:"GITFOX_REDIS_USE_SENTINEL"          default:"false"`
		SentinelMaster     string `envconfig:"GITFOX_REDIS_SENTINEL_MASTER"`
		SentinelEndpoint   string `envconfig:"GITFOX_REDIS_SENTINEL_ENDPOINT"`
	}

	Events struct {
		Mode                  events.Mode `envconfig:"GITFOX_EVENTS_MODE"                     default:"inmemory"`
		Namespace             string      `envconfig:"GITFOX_EVENTS_NAMESPACE"                default:"gitfox"`
		MaxStreamLength       int64       `envconfig:"GITFOX_EVENTS_MAX_STREAM_LENGTH"        default:"10000"`
		ApproxMaxStreamLength bool        `envconfig:"GITFOX_EVENTS_APPROX_MAX_STREAM_LENGTH" default:"true"`
	}

	Lock struct {
		// Provider is a name of distributed lock service like redis, memory, file etc...
		Provider      lock.Provider `envconfig:"GITFOX_LOCK_PROVIDER"          default:"inmemory"`
		Expiry        time.Duration `envconfig:"GITFOX_LOCK_EXPIRE"            default:"8s"`
		Tries         int           `envconfig:"GITFOX_LOCK_TRIES"             default:"8"`
		RetryDelay    time.Duration `envconfig:"GITFOX_LOCK_RETRY_DELAY"       default:"250ms"`
		DriftFactor   float64       `envconfig:"GITFOX_LOCK_DRIFT_FACTOR"      default:"0.01"`
		TimeoutFactor float64       `envconfig:"GITFOX_LOCK_TIMEOUT_FACTOR"    default:"0.25"`
		// AppNamespace is just service app prefix to avoid conflicts on key definition
		AppNamespace string `envconfig:"GITFOX_LOCK_APP_NAMESPACE"     default:"gitfox"`
		// DefaultNamespace is when mutex doesn't specify custom namespace for their keys
		DefaultNamespace string `envconfig:"GITFOX_LOCK_DEFAULT_NAMESPACE" default:"default"`
	}

	PubSub struct {
		// Provider is a name of distributed lock service like redis, memory, file etc...
		Provider pubsub.Provider `envconfig:"GITFOX_PUBSUB_PROVIDER"                default:"inmemory"`
		// AppNamespace is just service app prefix to avoid conflicts on channel definition
		AppNamespace string `envconfig:"GITFOX_PUBSUB_APP_NAMESPACE"                default:"gitfox"`
		// DefaultNamespace is custom namespace for their channels
		DefaultNamespace string        `envconfig:"GITFOX_PUBSUB_DEFAULT_NAMESPACE" default:"default"`
		HealthInterval   time.Duration `envconfig:"GITFOX_PUBSUB_HEALTH_INTERVAL"   default:"3s"`
		SendTimeout      time.Duration `envconfig:"GITFOX_PUBSUB_SEND_TIMEOUT"      default:"60s"`
		ChannelSize      int           `envconfig:"GITFOX_PUBSUB_CHANNEL_SIZE"      default:"100"`
	}

	BackgroundJobs struct {
		// MaxRunning is maximum number of jobs that can be running at once.
		MaxRunning int `envconfig:"GITFOX_JOBS_MAX_RUNNING" default:"10"`

		// RetentionTime is the duration after which non-recurring,
		// finished and failed jobs will be purged from the DB.
		RetentionTime time.Duration `envconfig:"GITFOX_JOBS_RETENTION_TIME" default:"120h"` // 5 days
	}

	Webhook struct {
		// UserAgentIdentity specifies the identity used for the user agent header
		// IMPORTANT: do not include version.
		UserAgentIdentity string `envconfig:"GITFOX_WEBHOOK_USER_AGENT_IDENTITY" default:"Gitfox"`
		// HeaderIdentity specifies the identity used for headers in webhook calls (e.g. X-Gitfox-Trigger, ...).
		// NOTE: If no value is provided, the UserAgentIdentity will be used.
		HeaderIdentity      string `envconfig:"GITFOX_WEBHOOK_HEADER_IDENTITY"`
		Concurrency         int    `envconfig:"GITFOX_WEBHOOK_CONCURRENCY" default:"4"`
		MaxRetries          int    `envconfig:"GITFOX_WEBHOOK_MAX_RETRIES" default:"3"`
		AllowPrivateNetwork bool   `envconfig:"GITFOX_WEBHOOK_ALLOW_PRIVATE_NETWORK" default:"false"`
		AllowLoopback       bool   `envconfig:"GITFOX_WEBHOOK_ALLOW_LOOPBACK" default:"false"`
		// RetentionTime is the duration after which webhook executions will be purged from the DB.
		RetentionTime time.Duration `envconfig:"GITFOX_WEBHOOK_RETENTION_TIME" default:"168h"` // 7 days
		// InternalWebhooksURL is the url for webhooks which are marked as internal
		InternalWebhooksURL string `envconfig:"GITFOX_WEBHOOK_INTERNAL_WEBHOOKS_URL"`
	}

	Trigger struct {
		Concurrency int `envconfig:"GITFOX_TRIGGER_CONCURRENCY" default:"4"`
		MaxRetries  int `envconfig:"GITFOX_TRIGGER_MAX_RETRIES" default:"3"`
	}

	Metric struct {
		Enabled  bool   `envconfig:"GITFOX_METRIC_ENABLED" default:"true"`
		Endpoint string `envconfig:"GITFOX_METRIC_ENDPOINT" default:"https://stats.drone.ci/api/v1/gitfox"`
		Token    string `envconfig:"GITFOX_METRIC_TOKEN"`
	}

	RepoSize struct {
		Enabled     bool          `envconfig:"GITFOX_REPO_SIZE_ENABLED" default:"true"`
		CRON        string        `envconfig:"GITFOX_REPO_SIZE_CRON" default:"0 0 * * *"`
		MaxDuration time.Duration `envconfig:"GITFOX_REPO_SIZE_MAX_DURATION" default:"15m"`
		NumWorkers  int           `envconfig:"GITFOX_REPO_SIZE_NUM_WORKERS" default:"5"`
	}

	CodeOwners struct {
		FilePaths []string `envconfig:"GITFOX_CODEOWNERS_FILEPATH" default:"CODEOWNERS,.gitfox/CODEOWNERS"`
	}

	SMTP struct {
		Host     string `envconfig:"GITFOX_SMTP_HOST"`
		Port     int    `envconfig:"GITFOX_SMTP_PORT"`
		Username string `envconfig:"GITFOX_SMTP_USERNAME"`
		Password string `envconfig:"GITFOX_SMTP_PASSWORD"`
		FromMail string `envconfig:"GITFOX_SMTP_FROM_MAIL"`
		Insecure bool   `envconfig:"GITFOX_SMTP_INSECURE"`
	}

	Notification struct {
		MaxRetries  int `envconfig:"GITFOX_NOTIFICATION_MAX_RETRIES" default:"3"`
		Concurrency int `envconfig:"GITFOX_NOTIFICATION_CONCURRENCY" default:"4"`
	}

	KeywordSearch struct {
		Concurrency int `envconfig:"GITFOX_KEYWORD_SEARCH_CONCURRENCY" default:"4"`
		MaxRetries  int `envconfig:"GITFOX_KEYWORD_SEARCH_MAX_RETRIES" default:"3"`
	}

	Repos struct {
		// DeletedRetentionTime is the duration after which deleted repositories will be purged.
		DeletedRetentionTime time.Duration `envconfig:"GITFOX_REPOS_DELETED_RETENTION_TIME" default:"2160h"` // 90 days
	}

	Docker struct {
		// Host sets the url to the docker server.
		Host string `envconfig:"GITFOX_DOCKER_HOST"`
		// APIVersion sets the version of the API to reach, leave empty for latest.
		APIVersion string `envconfig:"GITFOX_DOCKER_API_VERSION"`
		// CertPath sets the path to load the TLS certificates from.
		CertPath string `envconfig:"GITFOX_DOCKER_CERT_PATH"`
		// TLSVerify enables or disables TLS verification, off by default.
		TLSVerify string `envconfig:"GITFOX_DOCKER_TLS_VERIFY"`
		// MachineHostName is the public host name of the machine on which the Docker.Host is running.
		// If not set, it parses the host from the URL.Base (e.g. localhost from http://localhost:3000).
		MachineHostName string `envconfig:"GITFOX_DOCKER_MACHINE_HOST_NAME"`
	}

	IDE struct {
		VSCodeWeb struct {
			// Port is the port on which the VSCode Web will be accessible.
			Port int `envconfig:"GITFOX_IDE_VSCODEWEB_PORT" default:"8089"`
		}

		VSCode struct {
			// Port is the port on which the SSH server for VSCode will be accessible.
			Port int `envconfig:"GITFOX_IDE_VSCODE_PORT" default:"8088"`
		}
	}

	Gitspace struct {
		// DefaultBaseImage is used to create the Gitspace when no devcontainer.json is absent or doesn't have image.
		DefaultBaseImage string `envconfig:"GITFOX_GITSPACE_DEFAULT_BASE_IMAGE" default:"mcr.microsoft.com/devcontainers/base:dev-ubuntu-24.04"` //nolint:lll

		Enable bool `envconfig:"GITFOX_GITSPACE_ENABLE" default:"false"`

		AgentPort int `envconfig:"GITFOX_GITSPACE_AGENT_PORT" default:"8083"`

		Events struct {
			Concurrency int `envconfig:"GITFOX_GITSPACE_EVENTS_CONCURRENCY" default:"4"`
			MaxRetries  int `envconfig:"GITFOX_GITSPACE_EVENTS_MAX_RETRIES" default:"3"`
		}
	}

	UI struct {
		ShowPlugin bool `envconfig:"GITFOX_UI_SHOW_PLUGIN" default:"true"`
	}

	Registry struct {
		Enable  bool `envconfig:"GITFOX_REGISTRY_ENABLED" default:"true"`
		Storage struct {
			// StorageType defines the type of storage to use for the registry. Options are: `filesystem`, `s3aws`
			StorageType string `envconfig:"GITFOX_REGISTRY_STORAGE_TYPE" default:"filesystem"`

			// FileSystemStorage defines the configuration for the filesystem storage if StorageType is `filesystem`.
			FileSystemStorage struct {
				MaxThreads    int    `envconfig:"GITFOX_REGISTRY_FILESYSTEM_MAX_THREADS" default:"100"`
				RootDirectory string `envconfig:"GITFOX_REGISTRY_FILESYSTEM_ROOT_DIRECTORY"`
			}

			// S3Storage defines the configuration for the S3 storage if StorageType is `s3aws`.
			S3Storage struct {
				AccessKey                   string `envconfig:"GITFOX_REGISTRY_S3_ACCESS_KEY"`
				SecretKey                   string `envconfig:"GITFOX_REGISTRY_S3_SECRET_KEY"`
				Region                      string `envconfig:"GITFOX_REGISTRY_S3_REGION"`
				RegionEndpoint              string `envconfig:"GITFOX_REGISTRY_S3_REGION_ENDPOINT"`
				ForcePathStyle              bool   `envconfig:"GITFOX_REGISTRY_S3_FORCE_PATH_STYLE" default:"true"`
				Accelerate                  bool   `envconfig:"GITFOX_REGISTRY_S3_ACCELERATED" default:"false"`
				Bucket                      string `envconfig:"GITFOX_REGISTRY_S3_BUCKET"`
				Encrypt                     bool   `envconfig:"GITFOX_REGISTRY_S3_ENCRYPT" default:"false"`
				KeyID                       string `envconfig:"GITFOX_REGISTRY_S3_KEY_ID"`
				Secure                      bool   `envconfig:"GITFOX_REGISTRY_S3_SECURE" default:"true"`
				V4Auth                      bool   `envconfig:"GITFOX_REGISTRY_S3_V4_AUTH" default:"true"`
				ChunkSize                   int    `envconfig:"GITFOX_REGISTRY_S3_CHUNK_SIZE" default:"10485760"`
				MultipartCopyChunkSize      int    `envconfig:"GITFOX_REGISTRY_S3_MULTIPART_COPY_CHUNK_SIZE" default:"33554432"`
				MultipartCopyMaxConcurrency int    `envconfig:"GITFOX_REGISTRY_S3_MULTIPART_COPY_MAX_CONCURRENCY" default:"100"`
				MultipartCopyThresholdSize  int    `envconfig:"GITFOX_REGISTRY_S3_MULTIPART_COPY_THRESHOLD_SIZE" default:"33554432"` //nolint:lll
				RootDirectory               string `envconfig:"GITFOX_REGISTRY_S3_ROOT_DIRECTORY"`
				UseDualStack                bool   `envconfig:"GITFOX_REGISTRY_S3_USE_DUAL_STACK" default:"false"`
				LogLevel                    string `envconfig:"GITFOX_REGISTRY_S3_LOG_LEVEL" default:"info"`
				Delete                      bool   `envconfig:"GITFOX_REGISTRY_S3_DELETE_ENABLED" default:"true"`
				Redirect                    bool   `envconfig:"GITFOX_REGISTRY_S3_STORAGE_REDIRECT" default:"false"`
			}
		}

		HTTP struct {
			// GITFOX_REGISTRY_HTTP_SECRET is used to encrypt the upload session details during docker push.
			// If not provided, a random secret will be generated. This may cause problems with uploads if multiple
			// registries are behind a load-balancer
			Secret string `envconfig:"GITFOX_REGISTRY_HTTP_SECRET"`
		}

		//nolint:lll
		GarbageCollection struct {
			Enabled                     bool          `envconfig:"GITFOX_REGISTRY_GARBAGE_COLLECTION_ENABLED" default:"false"`
			NoIdleBackoff               bool          `envconfig:"GITFOX_REGISTRY_GARBAGE_COLLECTION_NO_IDLE_BACKOFF" default:"false"`
			MaxBackoffDuration          time.Duration `envconfig:"GITFOX_REGISTRY_GARBAGE_COLLECTION_MAX_BACKOFF_DURATION" default:"10m"`
			InitialIntervalDuration     time.Duration `envconfig:"GITFOX_REGISTRY_GARBAGE_COLLECTION_INITIAL_INTERVAL_DURATION" default:"5s"`     //nolint:lll
			TransactionTimeoutDuration  time.Duration `envconfig:"GITFOX_REGISTRY_GARBAGE_COLLECTION_TRANSACTION_TIMEOUT_DURATION" default:"10s"` //nolint:lll
			BlobsStorageTimeoutDuration time.Duration `envconfig:"GITFOX_REGISTRY_GARBAGE_COLLECTION_BLOB_STORAGE_TIMEOUT_DURATION" default:"5s"` //nolint:lll
		}
	}

	Instrumentation struct {
		Enable bool   `envconfig:"GITFOX_INSTRUMENTATION_ENABLE" default:"false"`
		Cron   string `envconfig:"GITFOX_INSTRUMENTATION_CRON" default:"0 0 * * *"`
	}
}
