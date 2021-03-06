package db

import (
	"context"
	"errors"

	"github.com/marccarre/kubernetes-deployment-strategies-workload/pkg/domain"
)

// SchemaVersion is the current version of the DB schema.
// N.B.: this constant should be updated every time new migrations are added.
const SchemaVersion = uint(2)

// DB is the interface for a database client.
type DB interface {
	// Ping ensures this database client can reach the database.
	Ping(ctx context.Context) error
	// CreateUser stores the provided user.
	CreateUser(ctx context.Context, user *domain.User) (int, error)
	// ReadUsers returns all stored users.
	ReadUsers(ctx context.Context) ([]*domain.User, error)
	// ReadUserByID return the stored user corresponding to the provided ID.
	ReadUserByID(ctx context.Context, id int) (*domain.User, error)
	// Close closes this connection to the database.
	Close() error
}

// ErrNotFound is returned when the requested user is not found.
var ErrNotFound = errors.New("not found")
