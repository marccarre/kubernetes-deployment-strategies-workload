// +build !integration

package dbtest

import (
	"testing"

	"github.com/stretchr/testify/assert" // More readable test assertions.

	"github.com/marccarre/kubernetes-deployment-strategies-workload/pkg/db"
)

// Setup sets up a new in-memory database.
func Setup(t *testing.T) db.DB {
	return NewInMemoryDB()
}

// Cleanup cleans up after a test.
func Cleanup(t *testing.T, database db.DB) {
	if database != nil {
		assert.NoError(t, database.Close())
	}
}
