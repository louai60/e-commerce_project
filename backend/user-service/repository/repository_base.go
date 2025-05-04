package repository

import (
	"context"
	"database/sql"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/louai60/e-commerce_project/backend/user-service/db"
	"go.uber.org/zap"
)

// RepositoryBase provides common functionality for all repositories
type RepositoryBase struct {
	dbConfig        *db.DBConfig
	logger          *zap.Logger
	replicaSelector db.ReplicaSelector
}

// NewRepositoryBase creates a new repository base
func NewRepositoryBase(dbConfig *db.DBConfig, logger *zap.Logger) *RepositoryBase {
	return &RepositoryBase{
		dbConfig:        dbConfig,
		logger:          logger,
		replicaSelector: db.RandomSelector(),
	}
}

// GetMaster returns the master database connection
func (r *RepositoryBase) GetMaster() *sqlx.DB {
	return r.dbConfig.Master
}

// GetReplica returns a replica database connection
// If no replicas are available, returns the master
func (r *RepositoryBase) GetReplica() *sqlx.DB {
	return r.dbConfig.GetReplicaOrMaster(r.replicaSelector)
}

// BeginTx starts a new transaction on the master database
func (r *RepositoryBase) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.dbConfig.Master.BeginTx(ctx, nil)
}

// ExecuteQuery executes a query on a replica if it's a read-only query,
// otherwise executes it on the master
func (r *RepositoryBase) ExecuteQuery(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// Simple heuristic to determine if this is a read-only query
	// In a real-world scenario, you might want to use a SQL parser
	isReadOnly := isReadOnlyQuery(query)

	if isReadOnly {
		return r.GetReplica().QueryContext(ctx, query, args...)
	}

	return r.GetMaster().QueryContext(ctx, query, args...)
}

// ExecuteQueryRow executes a query that returns a single row on a replica if it's a read-only query,
// otherwise executes it on the master
func (r *RepositoryBase) ExecuteQueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	isReadOnly := isReadOnlyQuery(query)

	if isReadOnly {
		return r.GetReplica().QueryRowContext(ctx, query, args...)
	}

	return r.GetMaster().QueryRowContext(ctx, query, args...)
}

// ExecuteExec executes a statement that doesn't return rows on the master
func (r *RepositoryBase) ExecuteExec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.GetMaster().ExecContext(ctx, query, args...)
}

// ExecuteNamedQuery executes a named query on a replica if it's a read-only query,
// otherwise executes it on the master
func (r *RepositoryBase) ExecuteNamedQuery(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error) {
	isReadOnly := isReadOnlyQuery(query)

	if isReadOnly {
		return r.GetReplica().NamedQueryContext(ctx, query, arg)
	}

	return r.GetMaster().NamedQueryContext(ctx, query, arg)
}

// ExecuteNamedExec executes a named statement that doesn't return rows on the master
func (r *RepositoryBase) ExecuteNamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return r.GetMaster().NamedExecContext(ctx, query, arg)
}

// isReadOnlyQuery determines if a query is read-only based on its first word
// This is a simple heuristic and might not work for all cases
func isReadOnlyQuery(query string) bool {
	// Convert to lowercase and trim spaces for case-insensitive comparison
	queryStart := ""
	for i := 0; i < len(query); i++ {
		if query[i] != ' ' && query[i] != '\n' && query[i] != '\t' {
			queryStart = query[i:]
			break
		}
	}

	// Check if the query starts with SELECT
	if len(queryStart) >= 6 && (strings.HasPrefix(strings.ToUpper(queryStart), "SELECT")) {
		return true
	}

	return false
}
