package config

import (
	"errors"
	"os"
)

type Server struct {
	Env  string `mapstructure:"env" envconfig:"APP_ENV" required:"true" default:""`
	Port string `envconfig:"PORT" required:"true" default:"8080" mapstructure:"-"`
}

type Test struct {
	Name string `mapstructure:"name" `
}

type DefaultConfig struct {
	Server Server `mapstructure:"server"`
	Test   Test   `mapstructure:"test"`
}

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
