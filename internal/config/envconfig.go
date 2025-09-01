package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

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
