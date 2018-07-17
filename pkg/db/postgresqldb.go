package db

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"                  // DB DSL.
	"github.com/golang-migrate/migrate"                   // DB migrations.
	"github.com/golang-migrate/migrate/database/postgres" // DB migrations for PostgreSQL.
	_ "github.com/golang-migrate/migrate/source/file"     // DB migrations for PostgreSQL.
	_ "github.com/lib/pq"                                 // DB PostgreSQL drivers.
	log "github.com/sirupsen/logrus"                      // Better Logging.

	"github.com/marccarre/kubernetes-deployment-strategies-workload/pkg/domain"
)

// PostgreSQLDB is a PostgreSQL-compatible implementation of DB.
type PostgreSQLDB struct {
	db *sql.DB
}

const driverName = "postgres"

// NewPostgreSQLDB creates a new connection to the configured PostgreSQL DB.
func NewPostgreSQLDB(config *Config) (*PostgreSQLDB, error) {
	uri, err := config.URI()
	if err != nil {
		log.WithField("err", err).Error("failed to get DB URI")
		return nil, err
	}
	db, err := sql.Open(driverName, uri)
	if err != nil {
		log.WithField("uri", uri).WithField("err", err).Error("failed to open connection")
		return nil, err
	}
	if err := runDBMigrations(db, config.MigrationsDir, config.SchemaVersion, uri); err != nil {
		return nil, err
	}
	return &PostgreSQLDB{
		db: db,
	}, nil
}

func runDBMigrations(db *sql.DB, migrationsDir string, targetVersion uint, uri string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.WithField("err", err).Error("failed to create DB migrations driver")
		return err
	}

	migrateClient, err := migrate.NewWithDatabaseInstance("file://"+migrationsDir, "users", driver)
	if err != nil {
		log.WithField("err", err).Error("failed to create DB migrations client")
		return err
	}
	defer migrateClient.Close()
	return checkOrUpdateSchema(migrateClient, targetVersion)
}

func checkOrUpdateSchema(migrateClient *migrate.Migrate, targetVersion uint) error {
	currentVersion, _, err := migrateClient.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.WithField("err", err).Error("failed to read DB's schema version")
		return err
	}
	logger := log.WithField("currentVersion", currentVersion).WithField("targetVersion", targetVersion)
	if currentVersion == targetVersion {
		logger.Info("DB already at the target schema version")
	} else {
		if currentVersion < targetVersion {
			logger.Info("upgrading DB schema...")
			err = migrateClient.Up()
		} else {
			logger.Info("downgrading DB schema...")
			err = migrateClient.Down()
		}
		if err != nil {
			if err == migrate.ErrNoChange {
				logger.WithField("msg", err).Info("DB already at the target schema version")
			} else {
				logger.WithField("err", err).Error("failed to apply DB migrations")
				return err
			}
		}
	}
	return nil
}

// IMPORTANT: make sure these match the latest migration under pkg/db/migrations/
const (
	users      = "users"
	id         = "id"
	firstName  = "first_name"
	familyName = "family_name"
)

// Ping ensures this database client can reach the database.
func (db PostgreSQLDB) Ping(ctx context.Context) error {
	return db.db.PingContext(ctx)
}

// CreateUser stores the provided user.
func (db PostgreSQLDB) CreateUser(ctx context.Context, user *domain.User) (int, error) {
	var id int
	err := debugInsert(
		db.query().
			Insert(users).
			Columns(firstName, familyName).
			Values(user.FirstName, user.FamilyName).
			Suffix("RETURNING id")).
		QueryRowContext(ctx).
		Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (db PostgreSQLDB) selectUsers() sq.SelectBuilder {
	// The order of the below columns ought to match
	// the order of the fields in scanUser and scanOne:
	return db.query().Select(id, firstName, familyName).From(users)
}

// ReadUsers returns all stored users.
func (db PostgreSQLDB) ReadUsers(ctx context.Context) ([]*domain.User, error) {
	rows, err := debugSelect(
		db.selectUsers().OrderBy("id ASC")).
		QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users, err := scanUsers(rows)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// ReadUserByID return the stored user corresponding to the provided ID.
func (db PostgreSQLDB) ReadUserByID(ctx context.Context, userID int) (*domain.User, error) {
	user, err := scanUser(debugSelect(
		db.selectUsers().Where(sq.Eq{id: userID})).
		QueryRowContext(ctx))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		} else {
			return nil, err
		}
	}
	return user, nil
}

func (db PostgreSQLDB) query() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db.db)
}

func debugInsert(query sq.InsertBuilder) sq.InsertBuilder {
	sql, args, err := query.ToSql()
	log.WithField("sql", sql).WithField("args", args).WithField("err", err).Debug("insert query")
	return query
}

func debugSelect(query sq.SelectBuilder) sq.SelectBuilder {
	sql, args, err := query.ToSql()
	log.WithField("sql", sql).WithField("args", args).WithField("err", err).Debug("select query")
	return query
}

func scanUsers(rows *sql.Rows) ([]*domain.User, error) {
	users := []*domain.User{}
	for rows.Next() {
		user, err := scanOne(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	err := rows.Err()
	if err != nil {
		return nil, err
	}
	return users, nil
}

func scanOne(rows *sql.Rows) (*domain.User, error) {
	user := &domain.User{}
	// The order of the below fields ought to match
	// the order of the columns in selectUsers:
	if err := rows.Scan(
		&user.ID,
		&user.FirstName,
		&user.FamilyName,
	); err != nil {
		return nil, err
	}
	return user, nil
}

func scanUser(row sq.RowScanner) (*domain.User, error) {
	user := &domain.User{}
	// The order of the below fields ought to match
	// the order of the columns in selectUsers:
	if err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.FamilyName,
	); err != nil {
		return nil, err
	}
	return user, nil
}

// Close closes this connection to the database.
func (db *PostgreSQLDB) Close() error {
	return db.db.Close()
}
