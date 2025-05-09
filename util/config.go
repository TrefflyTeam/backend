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
