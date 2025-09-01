package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
)

var current DefaultConfig

func Load() (cfg DefaultConfig, err error) {

	env := os.Getenv("APP_ENV")

	// 1) 파일 로드
	if err := loadAppConfig(&current, env); err != nil {
		if errors.Is(err, ErrBaseConfigNotFound) {
			// 경고로만 취급 → 호출자가 로그 레벨 결정
			// 로그 출력
		} else {
			// 치명 에러는 바로 반환
			return DefaultConfig{}, err
		}
	}

	// 2) ENV 오버레이
	if err := loadEnvConfig(&current); err != nil {
		return DefaultConfig{}, err
	}

	return current, nil
}

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

func loadEnvConfig(settings *DefaultConfig) error {
	prefix := ""

	if err := envconfig.Process(prefix, settings); err != nil {

		return fmt.Errorf("config: env overlay: %w", err)
	}

	return nil
}

// stdout으로 사용법 출력
func PrintUsage() {
	_ = envconfig.Usage("", &DefaultConfig{})
}