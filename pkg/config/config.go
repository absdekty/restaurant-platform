package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

/* Переиспользуемая конфигурация */
type AppConfig struct {
	ENV      string `mapstructure:"env"       validate:"required,oneof=dev prod"`
	AuthAddr string `mapstructure:"auth_addr" validate:"required,hostname_port"`
	UserAddr string `mapstructure:"user_addr" validate:"required,hostname_port"`
	CACert   string `mapstructure:"ca_cert"   validate:"required,filepath"`
}

/* REST Server */
type RESTConfig struct {
	Addr            string        `mapstructure:"addr"             validate:"required,hostname_port"`
	TimeoutRead     time.Duration `mapstructure:"timeout_read"     validate:"required,min=1s"`
	TimeoutWrite    time.Duration `mapstructure:"timeout_write"    validate:"required,min=1s"`
	TimeoutIdle     time.Duration `mapstructure:"timeout_idle"     validate:"required,min=1s"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" validate:"required,min=1s"`
}

/* gRPC Server */
type GRPCServerConfig struct {
	MaxRecvMsgSize   int           `mapstructure:"grpc_max_recv_msg_size" validate:"required,min=1048576"`
	MaxSendMsgSize   int           `mapstructure:"grpc_max_send_msg_size" validate:"required,min=1048576"`
	ConnTimeout      time.Duration `mapstructure:"grpc_conn_timeout"        validate:"required,min=1s"`
	MaxConnIdle      time.Duration `mapstructure:"grpc_max_conn_idle"       validate:"required,min=1s"`
	KeepaliveTime    time.Duration `mapstructure:"grpc_keepalive_time"      validate:"required,min=1s"`
	KeepaliveTimeout time.Duration `mapstructure:"grpc_keepalive_timeout"   validate:"required,min=1s"`
}

/* gRPC Client */
type GRPCClientConfig struct {
	RetryMaxAttempts       int           `mapstructure:"retry_max_attempts"        validate:"required,min=1"`
	RetryInitialBackoff    string        `mapstructure:"retry_initial_backoff"     validate:"required"`
	RetryMaxBackoff        string        `mapstructure:"retry_max_backoff"         validate:"required"`
	RetryBackoffMultiplier float64       `mapstructure:"retry_backoff_multiplier"  validate:"required,min=1"`
	KeepaliveTime          time.Duration `mapstructure:"keepalive_time"            validate:"required,min=1s"`
	KeepaliveTimeout       time.Duration `mapstructure:"keepalive_timeout"         validate:"required,min=1s"`
	KeepalivePermitWithout bool          `mapstructure:"keepalive_permit_without_stream"`
}

/* Redis Client */
type RedisClientConfig struct {
	Addr     string `mapstructure:"addr" validate:"required,hostname_port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db" validate:"min=0"`
	PoolSize int    `mapstructure:"poolsize" validate:"required,min=1"`
}

/* Circuit breaker */
type CircuitBreakerConfig struct {
	Name        string        `mapstructure:"name" validate:"required,min=1"`
	MaxRequests uint32        `mapstructure:"maxrequests" validate:"required,min=1"`
	Interval    time.Duration `mapstructure:"interval" validate:"min=5s"`
	Timeout     time.Duration `mapstructure:"timeout" validate:"required,min=10s"`
	MaxFailures uint32        `mapstructure:"maxfailures" validate:"required,min=1"`
}

/* Certs mTLS */
type CertsConfig struct {
	Cert    string `mapstructure:"cert" validate:"required,filepath"`
	CertKey string `mapstructure:"cert_key" validate:"required,filepath"`
}

/* PostgreSQL */
type PostgresConfig struct {
	Addr     string `mapstructure:"addr" validate:"required,hostname_port"`
	User     string `mapstructure:"user" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	Name     string `mapstructure:"db" validate:"required"`
	SSLMode  string `mapstructure:"ssl" validate:"required,oneof=disable require"`
}

/* =========== GATEWAY =========== */
/* Gateway */
type GatewayConfig struct {
	AppConfig `mapstructure:",squash"`
	Gateway   GatewaySettings `mapstructure:"gateway" validate:"required"`
}

type RateLimiterRule struct {
	Limits  map[string]int           `mapstructure:"limits" validate:"required"`
	Expires map[string]time.Duration `mapstructure:"expires" validate:"required"`
}

type RateLimiter struct {
	RedisClient RedisClientConfig `mapstructure:"redis_client" validate:"required"`
	All         RateLimiterRule   `mapstructure:"all" validate:"required"`
	IP          RateLimiterRule   `mapstructure:"ip" validate:"required"`
}

type GatewayCircuitBreakers struct {
	CBAuth CircuitBreakerConfig `mapstructure:"auth" validate:"required"`
	CBUser CircuitBreakerConfig `mapstructure:"user" validate:"required"`
}

type GatewaySettings struct {
	Server         RESTConfig             `mapstructure:"server" validate:"required"`
	CertsClient    CertsConfig            `mapstructure:"certs_client" validate:"required"`
	RateLimiter    RateLimiter            `mapstructure:"rate_limiter" validate:"required"`
	GRPCAuthClient GRPCClientConfig       `mapstructure:"grpc_client_auth"  validate:"required"`
	GRPCUserClient GRPCClientConfig       `mapstructure:"grpc_client_user"  validate:"required"`
	MetricsServer  RESTConfig             `mapstructure:"metrics"  validate:"required"`
	CircuitBreaker GatewayCircuitBreakers `mapstructure:"circuit_breakers" validate:"required"`
}

/* =========== AUTH =========== */
/* Auth */
type AuthConfig struct {
	AppConfig `mapstructure:",squash"`
	Auth      AuthSettings `mapstructure:"auth" validate:"required"`
}

type JWTConfig struct {
	SecretKey  string        `mapstructure:"secret_key"   validate:"required,min=32"`
	AccessTTL  time.Duration `mapstructure:"access_ttl"   validate:"required,min=15m"`
	RefreshTTL time.Duration `mapstructure:"refresh_ttl"  validate:"required,min=168h"`
}

type AuthSettings struct {
	ShutdownTimeout time.Duration     `mapstructure:"shutdown_timeout" validate:"required,min=1s"`
	CertsServer     CertsConfig       `mapstructure:"certs_server" validate:"required"`
	JWTConfig       JWTConfig         `mapstructure:"jwt"   validate:"required"`
	GRPCServer      GRPCServerConfig  `mapstructure:"grpc_server" validate:"required"`
	RedisClient     RedisClientConfig `mapstructure:"redis_client" validate:"required"`
}

/* =========== USER =========== */
/* User */
type UserConfig struct {
	AppConfig `mapstructure:",squash"`
	User      UserSettings `mapstructure:"user" validate:"required"`
}

type UserSettings struct {
	ShutdownTimeout time.Duration    `mapstructure:"shutdown_timeout" validate:"required,min=1s"`
	CertsServer     CertsConfig      `mapstructure:"certs_server" validate:"required"`
	CertsClient     CertsConfig      `mapstructure:"certs_client" validate:"required"`
	GRPCServer      GRPCServerConfig `mapstructure:"grpc_server" validate:"required"`
	PostgreSQL      PostgresConfig   `mapstructure:"postgres" validate:"required"`
}

/* =========== Загрузка конфига =========== */
func Load(path, envPrefix string, cfg any) error {
	v := viper.New()
	v.SetConfigType("yaml")

	// .yaml
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	expandedData := os.ExpandEnv(string(data))
	if err := v.ReadConfig(strings.NewReader(expandedData)); err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	// .local.yaml
	localPath := strings.TrimSuffix(path, ".yaml") + ".local.yaml"
	if _, err := os.Stat(localPath); err == nil {
		localData, err := os.ReadFile(localPath)
		if err != nil {
			return fmt.Errorf("read local config: %w", err)
		}
		expandedLocal := os.ExpandEnv(string(localData))
		if err := v.MergeConfig(strings.NewReader(expandedLocal)); err != nil {
			return fmt.Errorf("merge local config: %w", err)
		}
	}

	// ENV
	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Unmarshal
	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	// Валидация
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			first := validationErrors[0]
			return fmt.Errorf(
				"validation: field '%s' failed on '%s' (got '%v')",
				first.StructField(),
				first.Tag(),
				first.Value(),
			)
		}
		return fmt.Errorf("validation: %w", err)
	}

	return nil
}
