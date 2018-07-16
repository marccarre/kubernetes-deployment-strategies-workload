package db_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	flag "github.com/spf13/pflag"        // POSIX/GNU-style CLI arguments.
	"github.com/stretchr/testify/assert" // More readable test assertions.

	"github.com/marccarre/kubernetes-deployment-strategies-workload/pkg/db"
)

func TestParsingEmptyArgumentsShouldReturnDefaultConfig(t *testing.T) {
	config := parseArgs(t, []string{})
	assert.NotNil(t, config)

	uri, err := config.URI()
	assert.NoError(t, err)
	assert.Equal(t, "postgres://postgres@localhost:5432/users?sslmode=disable", uri)
	assert.Equal(t, "/home/service/migrations", config.MigrationsDir)
	assert.Equal(t, uint(1), config.SchemaVersion)
}

func TestParsingArgumentsShouldOverrideDefaultConfig(t *testing.T) {
	passwordFile, err := passwordFile("s3cr3t")
	assert.NoError(t, err)
	defer os.Remove(passwordFile.Name()) // Clean-up.

	migrationsDir, err := migrationsDir()
	assert.NoError(t, err)
	defer os.RemoveAll(migrationsDir) // Clean-up.

	config := parseArgs(t, []string{
		"--db-uri", "postgres://foo@bar:1337/baz?sslmode=verify-full",
		"--db-passwd-file", passwordFile.Name(),
		"--db-migrations-dir", migrationsDir,
		"--db-schema-version", "42",
	})
	assert.NotNil(t, config)

	uri, err := config.URI()
	assert.NoError(t, err)
	assert.Equal(t, "postgres://foo:s3cr3t@bar:1337/baz?sslmode=verify-full", uri)
	assert.Equal(t, migrationsDir, config.MigrationsDir)
	assert.Equal(t, uint(42), config.SchemaVersion)
}

func TestInvalidDatabaseURIShouldReturnError(t *testing.T) {
	config := parseArgs(t, []string{
		"--db-uri", "::not-a-valid-URI",
	})
	assert.NotNil(t, config)

	uri, err := config.URI()
	assert.EqualError(t, err, "failed to parse database URI: parse ::not-a-valid-URI: missing protocol scheme")
	assert.Equal(t, "", uri)
}

func TestPasswordFileWithMissingDatabaseUsernameInURIShouldReturnError(t *testing.T) {
	passwordFile, err := passwordFile("s3cr3t")
	assert.NoError(t, err)
	defer os.Remove(passwordFile.Name()) // Clean-up.

	config := parseArgs(t, []string{
		"--db-uri", "postgres://bar:1337/baz?sslmode=verify-full",
		"--db-passwd-file", passwordFile.Name(),
	})
	assert.NotNil(t, config)

	uri, err := config.URI()
	assert.EqualError(t, err, "invalid username: --db-passwd-file requires username in --db-uri but none provided in postgres://bar:1337/baz?sslmode=verify-full")
	assert.Equal(t, "", uri)
}

func TestInvalidDatabasePasswordFileShouldReturnError(t *testing.T) {
	passwordFile, err := passwordFile("unused-password")
	assert.NoError(t, err)
	defer os.Remove(passwordFile.Name()) // Clean-up.

	config := parseArgs(t, []string{
		"--db-passwd-file", passwordFile.Name(),
	})
	assert.NotNil(t, config)

	// Force removal of the password file before we try to read it, in order to trigger the error condition:
	os.Remove(passwordFile.Name())

	uri, err := config.URI()
	assert.EqualError(t, err, fmt.Sprintf("failed to read database password file: open %v: no such file or directory", passwordFile.Name()))
	assert.Equal(t, "", uri)
}

// Utility function to create a Config object, register CLI arguments, and parse them.
func parseArgs(t *testing.T, args []string) *db.Config {
	config := db.Config{}
	cli := flag.NewFlagSet("service-test", flag.ContinueOnError)
	config.RegisterFlags(cli)
	err := cli.Parse(args)
	assert.NoError(t, err)
	return &config
}

// Utility function to generate a temporary file with the provided password.
func passwordFile(password string) (*os.File, error) {
	passwordFile, err := ioutil.TempFile(os.TempDir(), "password")
	if err != nil {
		return nil, err
	}
	if _, err := passwordFile.Write([]byte(password)); err != nil {
		return nil, err
	}
	if err := passwordFile.Close(); err != nil {
		return nil, err
	}
	return passwordFile, nil
}

// Utility function to generate a temporary directory.
func migrationsDir() (string, error) {
	return ioutil.TempDir(os.TempDir(), "migrations")
}
