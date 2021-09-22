package config

import (
	"encoding/json"
	"sync"
)

// Config of the service
type Config struct {
	sync.RWMutex

	Debug       bool   `default:"false"       env:"DEBUG"`
	ServiceName string `default:"apptemplate" env:"SERVICE_NAME"`
	Listen      string `default:":80"         env:"SERVER_LISTEN"`
	ListenPprof string `default:":8080"       env:"SERVER_LISTEN_PPROF"`

	Logger struct {
		Level   string `default:"error"                        env:"LOGGER_LEVEL"`
		Address string `default:"fluentd.service.consul:24224" env:"LOGGER_ADDR"`
		Output  string `default:"stdout"                       env:"LOGGER_OUTPUT"`
	}

	Sentry struct {
		DSN              string  `default:""      env:"SENTRY_DSN"`
		Debug            bool    `default:"false" env:"SENTRY_DEBUG"`
		AttachStacktrace bool    `default:"true"  env:"SENTRY_ENABLE_STACK_TRACE"`
		SampleRate       float64 `default:"1.0"   env:"SENTRY_SAMPLE_RATE"`
	}
}

// String implementation of Stringer interface
func (c *Config) String() string {
	if out, err := json.MarshalIndent(&c, "", "  "); err == nil {
		return string(out)
	}

	return ""
}
