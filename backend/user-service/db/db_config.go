package db

import (
	// "database/sql"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/louai60/e-commerce_project/backend/user-service/config"
	"go.uber.org/zap"
)

// DBConfig holds the database configuration and connections
type DBConfig struct {
	Master   *sqlx.DB
	Replicas []*sqlx.DB
	mu       sync.Mutex
}

// ReplicaSelector is a function type that selects a replica from the available replicas
type ReplicaSelector func(replicas []*sqlx.DB) *sqlx.DB

// RoundRobinSelector returns a selector that selects replicas in a round-robin fashion
func RoundRobinSelector() ReplicaSelector {
	var currentIndex int
	var mu sync.Mutex

	return func(replicas []*sqlx.DB) *sqlx.DB {
		if len(replicas) == 0 {
			return nil
		}

		mu.Lock()
		defer mu.Unlock()

		selected := replicas[currentIndex]
		currentIndex = (currentIndex + 1) % len(replicas)
		return selected
	}
}

// RandomSelector returns a selector that selects replicas randomly
func RandomSelector() ReplicaSelector {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return func(replicas []*sqlx.DB) *sqlx.DB {
		if len(replicas) == 0 {
			return nil
		}

		return replicas[r.Intn(len(replicas))]
	}
}

// NewDBConfig creates a new database configuration with master and replicas
func NewDBConfig(cfg *config.Config, logger *zap.Logger) (*DBConfig, error) {
	// Connect to master
	masterDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
	)

	master, err := sqlx.Connect("postgres", masterDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master database: %w", err)
	}

	// Set connection pool settings
	master.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	master.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	master.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime.Minutes()) * time.Minute)

	logger.Info("Connected to master database",
		zap.String("host", cfg.Database.Host),
		zap.String("port", cfg.Database.Port),
		zap.String("database", cfg.Database.Name),
	)

	// Initialize replicas if configured
	var replicas []*sqlx.DB
	for i, replica := range cfg.Database.Replicas {
		replicaDSN := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			replica.Host,
			replica.Port,
			replica.User,
			replica.Password,
			replica.Name,
		)

		replicaDB, err := sqlx.Connect("postgres", replicaDSN)
		if err != nil {
			logger.Warn("Failed to connect to replica database",
				zap.String("host", replica.Host),
				zap.String("port", replica.Port),
				zap.String("database", replica.Name),
				zap.Error(err),
			)
			continue
		}

		// Set connection pool settings
		replicaDB.SetMaxOpenConns(replica.MaxOpenConns)
		replicaDB.SetMaxIdleConns(replica.MaxIdleConns)
		replicaDB.SetConnMaxLifetime(time.Duration(replica.ConnMaxLifetime.Minutes()) * time.Minute)

		logger.Info("Connected to replica database",
			zap.Int("index", i),
			zap.String("host", replica.Host),
			zap.String("port", replica.Port),
			zap.String("database", replica.Name),
		)

		replicas = append(replicas, replicaDB)
	}

	return &DBConfig{
		Master:   master,
		Replicas: replicas,
	}, nil
}

// GetReplicaOrMaster returns a replica if available, otherwise returns the master
func (c *DBConfig) GetReplicaOrMaster(selector ReplicaSelector) *sqlx.DB {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.Replicas) == 0 {
		return c.Master
	}

	replica := selector(c.Replicas)
	if replica == nil {
		return c.Master
	}

	return replica
}

// Close closes all database connections
func (c *DBConfig) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Master != nil {
		c.Master.Close()
	}

	for _, replica := range c.Replicas {
		replica.Close()
	}
}
