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

/* Gateway */
type GatewayConfig struct {
	AppConfig `mapstructure:",squash"`
	Gateway   GatewaySettings `mapstructure:"gateway" validate:"required"`
}

type GatewaySettings struct {
	Addr            string        `mapstructure:"addr"             validate:"required,hostname_port"`
	TimeoutRead     time.Duration `mapstructure:"timeout_read"     validate:"required,min=1s"`
	TimeoutWrite    time.Duration `mapstructure:"timeout_write"    validate:"required,min=1s"`
	TimeoutIdle     time.Duration `mapstructure:"timeout_idle"     validate:"required,min=1s"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" validate:"required,min=1s"`
	Cert            string        `mapstructure:"cert"             validate:"required,filepath"`
	CertKey         string        `mapstructure:"cert_key"         validate:"required,filepath"`
	RPS             int           `mapstructure:"rps"              validate:"required,min=1"`
	Burst           int           `mapstructure:"rps_burst"        validate:"required,min=1"`
}

/* Auth */
type AuthConfig struct {
	AppConfig `mapstructure:",squash"`
	Auth      AuthSettings `mapstructure:"auth" validate:"required"`
}

type AuthSettings struct {
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" validate:"required,min=1s"`
	Cert            string        `mapstructure:"cert"             validate:"required,filepath"`
	CertKey         string        `mapstructure:"cert_key"         validate:"required,filepath"`
	SecretKey       string        `mapstructure:"secret_key"   validate:"required,min=32"`
	AccessTTL       time.Duration `mapstructure:"access_ttl"   validate:"required,min=15m"`
	RefreshTTL      time.Duration `mapstructure:"refresh_ttl"  validate:"required,min=168h"`
}

/* User */
type UserConfig struct {
	AppConfig `mapstructure:",squash"`
	User      UserSettings `mapstructure:"user" validate:"required"`
}

type UserSettings struct {
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" validate:"required,min=1s"`
	Cert            string        `mapstructure:"cert"             validate:"required,filepath"`
	CertKey         string        `mapstructure:"cert_key"         validate:"required,filepath"`
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
