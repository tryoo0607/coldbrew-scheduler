package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func loadAppConfig(settings *DefaultConfig, appEnv string) error {
	path := "./configs"

	viper.Reset()
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// base 파일 없으면 경고성 에러로 전달
		return fmt.Errorf("%w: %v", ErrBaseConfigNotFound, err)
	}

	if appEnv != "" {
		viper.SetConfigName(fmt.Sprintf("app.%s", appEnv))
		viper.MergeInConfig()
	}

	if err := viper.Unmarshal(settings); err != nil {
		return fmt.Errorf("config: unmarshal yaml: %w", err)
	}

	return nil
}
