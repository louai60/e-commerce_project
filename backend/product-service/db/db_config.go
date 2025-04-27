package db

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/louai60/e-commerce_project/backend/product-service/config"
	"go.uber.org/zap"
)

// DBConfig holds the configuration for database connections
type DBConfig struct {
	Master   *sql.DB
	Replicas []*sql.DB
	Logger   *zap.Logger
}

// ReplicaSelector is a function type that selects a replica from a list
type ReplicaSelector func([]*sql.DB) *sql.DB

// RoundRobinSelector selects replicas in a round-robin fashion
func RoundRobinSelector() ReplicaSelector {
	var mu sync.Mutex
	var currentIndex int

	return func(replicas []*sql.DB) *sql.DB {
		if len(replicas) == 0 {
			return nil
		}

		mu.Lock()
		defer mu.Unlock()

		if currentIndex >= len(replicas) {
			currentIndex = 0
		}

		replica := replicas[currentIndex]
		currentIndex++
		return replica
	}
}

// RandomSelector selects replicas randomly
func RandomSelector() ReplicaSelector {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var mu sync.Mutex

	return func(replicas []*sql.DB) *sql.DB {
		if len(replicas) == 0 {
			return nil
		}

		mu.Lock()
		defer mu.Unlock()

		index := r.Intn(len(replicas))
		return replicas[index]
	}
}

// NewDBConfig creates a new database configuration with master and replicas
func NewDBConfig(cfg *config.Config, logger *zap.Logger) (*DBConfig, error) {
	// Connect to master database
	masterDB, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master database: %w", err)
	}

	// Configure connection pool settings
	masterDB.SetMaxOpenConns(25)
	masterDB.SetMaxIdleConns(5)
	masterDB.SetConnMaxLifetime(5 * time.Minute)

	// Test master connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := masterDB.PingContext(ctx); err != nil {
		masterDB.Close()
		return nil, fmt.Errorf("failed to ping master database: %w", err)
	}

	// Initialize replicas array
	var replicas []*sql.DB

	// Connect to replica databases if configured
	for i, replicaConfig := range cfg.Database.Replicas {
		replicaDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			replicaConfig.Host,
			replicaConfig.Port,
			replicaConfig.User,
			cfg.Secrets.DatabasePassword, // Assuming same password for replicas
			replicaConfig.Name,
			replicaConfig.SSLMode,
		)

		replicaDB, err := sql.Open("postgres", replicaDSN)
		if err != nil {
			logger.Warn("Failed to connect to replica database",
				zap.Int("replica_index", i),
				zap.String("host", replicaConfig.Host),
				zap.Error(err))
			continue
		}

		// Configure connection pool settings for replica
		replicaDB.SetMaxOpenConns(25)
		replicaDB.SetMaxIdleConns(5)
		replicaDB.SetConnMaxLifetime(5 * time.Minute)

		// Test replica connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := replicaDB.PingContext(ctx); err != nil {
			cancel()
			replicaDB.Close()
			logger.Warn("Failed to ping replica database",
				zap.Int("replica_index", i),
				zap.String("host", replicaConfig.Host),
				zap.Error(err))
			continue
		}
		cancel()

		replicas = append(replicas, replicaDB)
		logger.Info("Connected to replica database",
			zap.Int("replica_index", i),
			zap.String("host", replicaConfig.Host))
	}

	dbConfig := &DBConfig{
		Master:   masterDB,
		Replicas: replicas,
		Logger:   logger,
	}

	// Initialize sharding if enabled
	if cfg.Database.Sharding.Enabled {
		logger.Info("Initializing database sharding",
			zap.String("strategy", cfg.Database.Sharding.Strategy),
			zap.Int("shardCount", cfg.Database.Sharding.ShardCount),
		)

		// Create shard configurations
		shardConfigs := make([]ShardConfig, len(cfg.Database.Sharding.Shards))
		for i, shard := range cfg.Database.Sharding.Shards {
			shardConfigs[i] = ShardConfig{
				ShardID:  shard.ShardID,
				Host:     shard.Host,
				Port:     shard.Port,
				User:     shard.User,
				Password: cfg.Secrets.DatabasePassword,
				DBName:   shard.Name,
			}
		}

		// Create sharding strategy
		var strategy ShardingStrategy
		switch cfg.Database.Sharding.Strategy {
		case "consistent_hashing":
			strategy = NewConsistentHashingStrategy(
				cfg.Database.Sharding.ShardCount,
				cfg.Database.Sharding.VirtualNodes,
			)
		case "modulo":
			fallthrough
		default:
			strategy = NewModuloShardingStrategy(
				cfg.Database.Sharding.ShardCount,
			)
		}

		// Create shard manager
		shardManager := NewShardManager(dbConfig, strategy)

		// Initialize shard manager
		if err := shardManager.Initialize(context.Background(), shardConfigs); err != nil {
			logger.Error("Failed to initialize shard manager", zap.Error(err))
			// Continue without sharding
		} else {
			logger.Info("Sharding initialized successfully")
		}
	}

	return dbConfig, nil
}

// Close closes all database connections
func (c *DBConfig) Close() {
	if c.Master != nil {
		c.Master.Close()
	}

	for i, replica := range c.Replicas {
		if replica != nil {
			if err := replica.Close(); err != nil {
				c.Logger.Warn("Failed to close replica connection",
					zap.Int("replica_index", i),
					zap.Error(err))
			}
		}
	}
}

// GetReplicaOrMaster returns a replica if available, otherwise returns the master
func (c *DBConfig) GetReplicaOrMaster(selector ReplicaSelector) *sql.DB {
	if len(c.Replicas) > 0 {
		if replica := selector(c.Replicas); replica != nil {
			return replica
		}
	}
	return c.Master
}
