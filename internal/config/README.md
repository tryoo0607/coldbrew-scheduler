### Application Property 파일
- app.yaml로 관리
- 환경변수에 따라 다음처럼 사용도 가능
	- app.dev.yaml
	- app.prod.yaml
	- app.test.yaml
- 환경변수에 따른 Property 분리 시 -> ENV 값 APP_ENV를 전달해 줘야 함
	- APP_ENV=dev

### 환경 변수 관리 방법
```go
type Settings struct {
	// envconfig만 쓰는 경우
	EnvConfigTest struct {
		Port string `envconfig: "PORT" required "true" default "8080"`
	} `mapstructure:"-"`

	// Viper만 쓰는 경우
	AppConfigTest struct {
		Test string `mapstructure:"TEST"`
	} `mapstructure:"viper-test"`

	// YAML 기본값 + ENV 오버라이드 둘 다 사용하는 경우
	MixedTest struct {
		Port int `mapstructure:"port" envconfig:"PORT" default:"8080"`
	} `mapstructure:"server"`
}
```

```go
type EnvConfigTest struct {
	Port string `envconfig: "PORT" required "true" default "8080"`
}

// Viper만 쓰는 경우
type AppConfigTest struct {
	Test string `mapstructure:"TEST"`
}

// YAML 기본값 + ENV 오버라이드 둘 다 사용하는 경우
type MixedTest struct {
	Port int `mapstructure:"port" envconfig:"PORT" default:"8080"`
}

type Settings struct {
	EnvConfigTest	EnvConfigTest	`mapstructure:"-"`
	AppConfigTest	AppConfigTest	`mapstructure:"viper-test"`
	MixedTest		MixedTest		`mapstructure:"server"`
}
```

### 환경 변수(Env)