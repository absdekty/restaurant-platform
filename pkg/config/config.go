package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

/* Общий конфиг */
type AppConfig struct {
	ENV      string `mapstructure:"env"       validate:"required,oneof=dev prod"`
	AuthAddr string `mapstructure:"auth_addr" validate:"required,hostname_port"`
	UserAddr string `mapstructure:"user_addr" validate:"required,hostname_port"`
	CACert   string `mapstructure:"ca_cert"   validate:"required,filepath"`
}

/* gRPC-Clients */
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

type MetricsServer struct {
	Addr            string        `mapstructure:"addr"             validate:"required,hostname_port"`
	TimeoutRead     time.Duration `mapstructure:"timeout_read"     validate:"required,min=1s"`
	TimeoutWrite    time.Duration `mapstructure:"timeout_write"    validate:"required,min=1s"`
	TimeoutIdle     time.Duration `mapstructure:"timeout_idle"     validate:"required,min=1s"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" validate:"required,min=1s"`
}

type GatewayCircuitBreakers struct {
	CBAuth CircuitBreakerConfig `mapstructure:"auth" validate:"required"`
	CBUser CircuitBreakerConfig `mapstructure:"auth" validate:"required"`
}

type GatewaySettings struct {
	Addr            string                 `mapstructure:"addr"             validate:"required,hostname_port"`
	TimeoutRead     time.Duration          `mapstructure:"timeout_read"     validate:"required,min=1s"`
	TimeoutWrite    time.Duration          `mapstructure:"timeout_write"    validate:"required,min=1s"`
	TimeoutIdle     time.Duration          `mapstructure:"timeout_idle"     validate:"required,min=1s"`
	ShutdownTimeout time.Duration          `mapstructure:"shutdown_timeout" validate:"required,min=1s"`
	CertsClient     CertsConfig            `mapstructure:"certs_client" validate:"required"`
	RateLimiter     RateLimiter            `mapstructure:"rate_limiter" validate:"required"`
	GRPCAuthClient  GRPCClientConfig       `mapstructure:"gRPCAuthClient"  validate:"required"`
	GRPCUserClient  GRPCClientConfig       `mapstructure:"gRPCUserClient"  validate:"required"`
	MetricsServer   MetricsServer          `mapstructure:"metrics"  validate:"required"`
	CircuitBreaker  GatewayCircuitBreakers `mapstructure:"circuit_breakers" validate:"required"`
}

/* Auth */
type AuthConfig struct {
	AppConfig `mapstructure:",squash"`
	Auth      AuthSettings `mapstructure:"auth" validate:"required"`
}

type AuthSettings struct {
	ShutdownTimeout      time.Duration     `mapstructure:"shutdown_timeout" validate:"required,min=1s"`
	CertsServer          CertsConfig       `mapstructure:"certs_server" validate:"required"`
	SecretKey            string            `mapstructure:"secret_key"   validate:"required,min=32"`
	AccessTTL            time.Duration     `mapstructure:"access_ttl"   validate:"required,min=15m"`
	RefreshTTL           time.Duration     `mapstructure:"refresh_ttl"  validate:"required,min=168h"`
	GRPCMaxRecvMsgSize   int               `mapstructure:"grpc_max_recv_msg_size" validate:"required,min=1048576"`
	GRPCMaxSendMsgSize   int               `mapstructure:"grpc_max_send_msg_size" validate:"required,min=1048576"`
	GRPCConnTimeout      time.Duration     `mapstructure:"grpc_conn_timeout"        validate:"required,min=1s"`
	GRPCMaxConnIdle      time.Duration     `mapstructure:"grpc_max_conn_idle"       validate:"required,min=1s"`
	GRPCKeepaliveTime    time.Duration     `mapstructure:"grpc_keepalive_time"      validate:"required,min=1s"`
	GRPCKeepaliveTimeout time.Duration     `mapstructure:"grpc_keepalive_timeout"   validate:"required,min=1s"`
	RedisClient          RedisClientConfig `mapstructure:"redis_client" validate:"required"`
}

/* User */
type UserConfig struct {
	AppConfig `mapstructure:",squash"`
	User      UserSettings `mapstructure:"user" validate:"required"`
}

type UserSettings struct {
	ShutdownTimeout      time.Duration    `mapstructure:"shutdown_timeout" validate:"required,min=1s"`
	CertsServer          CertsConfig      `mapstructure:"certs_server" validate:"required"`
	CertsClient          CertsConfig      `mapstructure:"certs_client" validate:"required"`
	GRPCMaxRecvMsgSize   int              `mapstructure:"grpc_max_recv_msg_size" validate:"required,min=1048576"`
	GRPCMaxSendMsgSize   int              `mapstructure:"grpc_max_send_msg_size" validate:"required,min=1048576"`
	GRPCConnTimeout      time.Duration    `mapstructure:"grpc_conn_timeout"        validate:"required,min=1s"`
	GRPCMaxConnIdle      time.Duration    `mapstructure:"grpc_max_conn_idle"       validate:"required,min=1s"`
	GRPCKeepaliveTime    time.Duration    `mapstructure:"grpc_keepalive_time"      validate:"required,min=1s"`
	GRPCKeepaliveTimeout time.Duration    `mapstructure:"grpc_keepalive_timeout"   validate:"required,min=1s"`
	GRPCAuthClient       GRPCClientConfig `mapstructure:"gRPCAuthClient"  validate:"required"`
	PostgreSQL           PostgresConfig   `mapstructure:"postgres" validate:"required"`
}

func Load(path, envPrefix string, cfg any) error {
	v := viper.New()

	// YAML
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("read YAML config: %w", err)
	}

	// Local YAML
	v.SetConfigFile(strings.TrimSuffix(path, ".yaml") + ".local.yaml")
	localPath := strings.TrimSuffix(path, ".yaml") + ".local.yaml"
	if _, err := os.Stat(localPath); err == nil {
		v.SetConfigFile(localPath)
		if err := v.MergeInConfig(); err != nil {
			return fmt.Errorf("merge local config: %w", err)
		}
	}

	// ENV
	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

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
