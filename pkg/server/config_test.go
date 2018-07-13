package server_test

import (
	"testing"
	"time"

	flag "github.com/spf13/pflag"        // POSIX/GNU-style CLI arguments.
	"github.com/stretchr/testify/assert" // More readable test assertions.

	"github.com/marccarre/kubernetes-deployment-strategies-workload/pkg/server"
)

func TestParsingEmptyArgumentsShouldReturnDefaultConfig(t *testing.T) {
	config := parseArgs(t, []string{})
	assert.NotNil(t, config)
	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, 15*time.Second, config.ReadTimeout)
	assert.Equal(t, 15*time.Second, config.WriteTimeout)
	assert.Equal(t, 60*time.Second, config.IdleTimeout)
}

func TestParsingArgumentsShouldOverrideDefaultConfig(t *testing.T) {
	config := parseArgs(t, []string{
		"--http-port", "1337",
		"--http-read-timeout", "10s",
		"--http-write-timeout", "20s",
		"--http-idle-timeout", "30s",
	})
	assert.NotNil(t, config)
	assert.Equal(t, 1337, config.Port)
	assert.Equal(t, 10*time.Second, config.ReadTimeout)
	assert.Equal(t, 20*time.Second, config.WriteTimeout)
	assert.Equal(t, 30*time.Second, config.IdleTimeout)
}

// Utility function to create a Config object, register CLI arguments, and parse them.
func parseArgs(t *testing.T, args []string) *server.Config {
	config := server.Config{}
	cli := flag.NewFlagSet("service-test", flag.ContinueOnError)
	config.RegisterFlags(cli)
	err := cli.Parse(args)
	assert.NoError(t, err)
	return &config
}
