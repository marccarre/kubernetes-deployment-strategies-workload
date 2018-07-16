package db

import (
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/pkg/errors"
	flag "github.com/spf13/pflag" // POSIX/GNU-style CLI arguments.
)

// Config encapsulates the input required to configure a database connection.
type Config struct {
	RawURI        string
	passwordFile  string
	MigrationsDir string
	SchemaVersion uint
}

const (
	dbURI           = "db-uri"
	dbMigrationsDir = "db-migrations-dir"
	dbSchemaVersion = "db-schema-version"
	dbPasswdFile    = "db-passwd-file"
)

// RegisterFlags maps the provided CLI arguments to fields in this configuration object.
func (cfg *Config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&cfg.RawURI, dbURI, "postgres://postgres@localhost:5432/users?sslmode=disable", "URI to connect to the database")
	f.StringVar(&cfg.passwordFile, dbPasswdFile, "", fmt.Sprintf("File containing the password to authenticate against the database (username goes in --%v)", dbURI))
	f.StringVar(&cfg.MigrationsDir, dbMigrationsDir, "/home/service/migrations", "Directory containing the database migrations to apply on application startup")
	f.UintVar(&cfg.SchemaVersion, dbSchemaVersion, SchemaVersion, "Version of the schema of the database. This version will be applied on application startup")
}

// URI parses this configuration object's database URI, reads the database password from the specified file, and injects it in the URI it returns.
func (cfg Config) URI() (string, error) {
	uri, err := url.Parse(cfg.RawURI)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse database URI")
	}

	if len(cfg.passwordFile) > 0 {
		if uri.User == nil {
			return "", fmt.Errorf("invalid username: --%v requires username in --%v but none provided in %v", dbPasswdFile, dbURI, cfg.RawURI)
		}
		passwordBytes, err := ioutil.ReadFile(cfg.passwordFile)
		if err != nil {
			return "", errors.Wrap(err, "failed to read database password file")
		}
		uri.User = url.UserPassword(uri.User.Username(), string(passwordBytes))
	}
	return uri.String(), nil
}
