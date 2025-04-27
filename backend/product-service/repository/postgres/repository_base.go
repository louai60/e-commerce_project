package postgres

import (
	"context"
	"database/sql"
	"sync"

	"github.com/louai60/e-commerce_project/backend/product-service/db"
	"github.com/louai60/e-commerce_project/backend/product-service/models"
	"go.uber.org/zap"
)

// RepositoryBase provides common functionality for all repositories
type RepositoryBase struct {
	dbConfig         *db.DBConfig
	logger           *zap.Logger
	replicaSelector  db.ReplicaSelector
	shardManager     *db.ShardManager
	shardKeyProvider models.ShardKeyProvider
	mu               sync.Mutex
	currentReplica   int
	shardingEnabled  bool
}

// NewRepositoryBase creates a new repository base
func NewRepositoryBase(dbConfig *db.DBConfig, logger *zap.Logger) *RepositoryBase {
	return &RepositoryBase{
		dbConfig:         dbConfig,
		logger:           logger,
		replicaSelector:  db.RandomSelector(),
		shardKeyProvider: models.DefaultShardKeyProvider,
		shardingEnabled:  false, // Disabled by default until explicitly initialized
	}
}

// InitSharding initializes sharding for the repository
func (r *RepositoryBase) InitSharding(shardManager *db.ShardManager, shardKeyProvider models.ShardKeyProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.shardManager = shardManager
	if shardKeyProvider != nil {
		r.shardKeyProvider = shardKeyProvider
	}
	r.shardingEnabled = true
}

// GetMaster returns the master database connection
func (r *RepositoryBase) GetMaster() *sql.DB {
	return r.dbConfig.Master
}

// GetReplica returns a replica database connection
// If no replicas are available, returns the master
func (r *RepositoryBase) GetReplica() *sql.DB {
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
	if len(queryStart) >= 6 && (queryStart[:6] == "SELECT" || queryStart[:6] == "select") {
		return true
	}

	return false
}

// GetDBForModel returns the appropriate database for a model based on sharding configuration
func (r *RepositoryBase) GetDBForModel(model interface{}) *sql.DB {
	if !r.shardingEnabled || r.shardManager == nil {
		return r.GetMaster()
	}

	// Get shard key from the model
	shardKey := r.shardKeyProvider(model)
	if shardKey == "" {
		return r.GetMaster()
	}

	// Get the appropriate shard
	shard := r.shardManager.GetShardForKey(shardKey)
	if shard == nil {
		return r.GetMaster()
	}

	return shard.Master
}

// GetReplicaForModel returns the appropriate replica for a model based on sharding configuration
func (r *RepositoryBase) GetReplicaForModel(model interface{}) *sql.DB {
	if !r.shardingEnabled || r.shardManager == nil {
		return r.GetReplica()
	}

	// Get shard key from the model
	shardKey := r.shardKeyProvider(model)
	if shardKey == "" {
		return r.GetReplica()
	}

	// Get the appropriate shard
	shard := r.shardManager.GetShardForKey(shardKey)
	if shard == nil {
		return r.GetReplica()
	}

	return shard.GetReplicaOrMaster(r.replicaSelector)
}

// ExecuteQueryForModel executes a query on the appropriate database for a model
func (r *RepositoryBase) ExecuteQueryForModel(ctx context.Context, model interface{}, query string, args ...interface{}) (*sql.Rows, error) {
	isReadOnly := isReadOnlyQuery(query)

	if isReadOnly {
		return r.GetReplicaForModel(model).QueryContext(ctx, query, args...)
	}

	return r.GetDBForModel(model).QueryContext(ctx, query, args...)
}

// ExecuteQueryRowForModel executes a query that returns a single row on the appropriate database for a model
func (r *RepositoryBase) ExecuteQueryRowForModel(ctx context.Context, model interface{}, query string, args ...interface{}) *sql.Row {
	isReadOnly := isReadOnlyQuery(query)

	if isReadOnly {
		return r.GetReplicaForModel(model).QueryRowContext(ctx, query, args...)
	}

	return r.GetDBForModel(model).QueryRowContext(ctx, query, args...)
}

// ExecuteExecForModel executes a statement that doesn't return rows on the appropriate database for a model
func (r *RepositoryBase) ExecuteExecForModel(ctx context.Context, model interface{}, query string, args ...interface{}) (sql.Result, error) {
	return r.GetDBForModel(model).ExecContext(ctx, query, args...)
}
