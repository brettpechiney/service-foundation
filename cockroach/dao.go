package cockroach

import (
	"database/sql"

	_ "github.com/lib/pq" // Import underlying database driver.
	"github.com/pkg/errors"
)

// DAO provides application-level context to the database handle.
type DAO struct {
	db *sql.DB
}

// Tx provides application-level context to the transaction handle.
type Tx struct {
	tx *sql.Tx
}

// NewDAO creates a database object, associates it with the
// Postgres driver, and validates the database connection.
func NewDAO(dataSource string) (DAO, error) {
	conn, err := sql.Open("postgres", dataSource)
	if err != nil {
		return DAO{}, errors.Wrap(err, "failed to open database connection")
	}
	return DAO{conn}, nil
}

// Close closes the underlying sql.DB
func (m DAO) Close() error {
	return m.db.Close()
}
