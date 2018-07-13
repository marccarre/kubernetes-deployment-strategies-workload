package server

import (
	"time"

	flag "github.com/spf13/pflag" // POSIX/GNU-style CLI arguments.
)

// Config encapsulates the input required to configure a database connection.
type Config struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

const (
	port         = "http-port"
	readTimeout  = "http-read-timeout"
	writeTimeout = "http-write-timeout"
	idleTimeout  = "http-idle-timeout"
)

// RegisterFlags maps the provided CLI arguments to fields in this configuration object.
func (cfg *Config) RegisterFlags(f *flag.FlagSet) {
	f.IntVar(&cfg.Port, port, 8080, "Port to connect to this web service")
	f.DurationVar(&cfg.ReadTimeout, readTimeout, 15*time.Second, "The maximum duration for reading the entire request, including the body.")
	f.DurationVar(&cfg.WriteTimeout, writeTimeout, 15*time.Second, "The maximum duration before timing out writes of the response.")
	f.DurationVar(&cfg.IdleTimeout, idleTimeout, 60*time.Second, "The maximum amount of time to wait for the next request when keep-alives are enabled.")
}
