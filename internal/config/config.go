package config

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