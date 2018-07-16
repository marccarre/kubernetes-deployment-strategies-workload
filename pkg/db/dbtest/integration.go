// +build integration

package dbtest

import (
	"testing"

	"github.com/stretchr/testify/assert" // More readable test assertions.

	"github.com/marccarre/kubernetes-deployment-strategies-workload/pkg/db"
)

// Setup sets up a new in-memory database.
func Setup(t *testing.T) db.DB {
	// The below values ought to match what is configured
	// in the Makefile, under the integration-test target:
	config := &db.Config{
		RawURI:        "postgres://postgres@users-db.local:5432/users_test?sslmode=disable",
		MigrationsDir: "/go/src/github.com/marccarre/kubernetes-deployment-strategies-workload/pkg/db/migrations",
		SchemaVersion: db.SchemaVersion,
	}
	database, err := db.NewPostgreSQLDB(config)
	assert.NoError(t, err)
	assert.NotNil(t, database)
	return database
}

// Cleanup cleans up after a test.
func Cleanup(t *testing.T, db db.DB) {
	if db != nil {
		assert.NoError(t, db.Close())
	}
}
