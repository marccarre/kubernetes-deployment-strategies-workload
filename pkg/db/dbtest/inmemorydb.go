package dbtest

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/marccarre/kubernetes-deployment-strategies-workload/pkg/db"
	"github.com/marccarre/kubernetes-deployment-strategies-workload/pkg/domain"
)

// InMemoryDB is an in-memory implementation of DB. This is mainly useful for testing.
type InMemoryDB struct {
	users  map[int]*domain.User
	nextID int
	mutex  sync.Mutex // For thread-safe access to the users map.
}

// NewInMemoryDB creates a new in-memory DB.
func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{
		users:  make(map[int]*domain.User),
		nextID: 1,
	}
}

// CreateUser stores the provided user.
func (database *InMemoryDB) CreateUser(_ context.Context, user *domain.User) (int, error) {
	database.mutex.Lock()
	defer database.mutex.Unlock()

	user.ID = max(user.ID, database.nextID)
	database.nextID = user.ID + 1

	if existingUser, ok := database.users[user.ID]; ok {
		return 0, fmt.Errorf("invalid user: ID already used by %v", *existingUser)
	}
	database.users[user.ID] = user
	return user.ID, nil
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

// ReadUsers returns all stored users.
func (database *InMemoryDB) ReadUsers(_ context.Context) ([]*domain.User, error) {
	database.mutex.Lock()
	defer database.mutex.Unlock()
	return toArray(database.users), nil
}

func toArray(usersMap map[int]*domain.User) []*domain.User {
	users := make([]*domain.User, len(usersMap))
	i := 0
	for _, user := range usersMap {
		users[i] = user
		i++
	}
	sort.Sort(ByID(users))
	return users
}

// ByID implements sort.Interface for []*domain.User based on the ID field.
type ByID []*domain.User

func (a ByID) Len() int           { return len(a) }
func (a ByID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByID) Less(i, j int) bool { return a[i].ID < a[j].ID }

// ReadUserByID return the stored user corresponding to the provided ID.
func (database *InMemoryDB) ReadUserByID(_ context.Context, id int) (*domain.User, error) {
	database.mutex.Lock()
	defer database.mutex.Unlock()
	if user, ok := database.users[id]; ok {
		return user, nil
	}
	return nil, db.ErrNotFound
}

// Close is a no-op, but present so that we implement the DB interface.
func (database *InMemoryDB) Close() error {
	return nil
}
