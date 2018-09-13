package jenkins_jr

import (
	"github.com/jmoiron/sqlx"
)

// Env contains global (not per-request) data.
type Env struct {
	DB *sqlx.DB
}
