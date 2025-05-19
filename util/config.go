package util

import (
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Environment           string        `mapstructure:"ENVIRONMENT"`
	DBDriver              string        `mapstructure:"DB_DRIVER"`
	DBSource              string        `mapstructure:"DB_SOURCE"`
	ServerAddress         string        `mapstructure:"SERVER_ADDRESS"`
	TokenSymmetricKey     string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration   time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration  time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	YandexGeocoderAPIKey  string        `mapstructure:"Y_GEOCODER_API"`
	YandexSuggesterAPIKey string        `mapstructure:"Y_SUGGESTER_API"`
	Domain                string        `mapstructure:"DOMAIN"`
	ImageBasePath         string        `mapstructure:"IMAGE_BASE_PATH"`
	GenBaseURL            string        `mapstructure:"GEN_BASE_URL"`
	GenAPIKey             string        `mapstructure:"GEN_API_KEY"`
	GenSystemPrompt       string        `mapstructure:"GEN_SYSTEM_PROMPT"`
	GenModel              string        `mapstructure:"GEN_MODEL"`
	RedisHost             string        `mapstructure:"REDIS_HOST"`
	RedisPort             int           `mapstructure:"REDIS_PORT"`
	RedisPassword         string        `mapstructure:"REDIS_PASSWORD"`
	RedisDB               int           `mapstructure:"REDIS_DB"`
	GenLimit              int           `mapstructure:"GEN_LIMIT"`
	GenTimeout            time.Duration `mapstructure:"GEN_TIMEOUT"`
	ResetCodeTTL          time.Duration `mapstructure:"RESET_CODE_TTL"`
	ResetCodeLength       int           `mapstructure:"RESET_CODE_LENGTH"`
	ResetTokenDuration    time.Duration `mapstructure:"RESET_TOKEN_DURATION"`
	SMTPHost              string        `mapstructure:"SMTP_HOST"`
	SMTPPort              int           `mapstructure:"SMTP_PORT"`
	SMTPUsername          string        `mapstructure:"SMTP_USERNAME"`
	SMTPPassword          string        `mapstructure:"SMTP_PASSWORD"`
	SMTPDefaultFrom       string        `mapstructure:"SMTP_DEFAULT_FROM"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.SetDefault("ENVIRONMENT", "development")

	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}
