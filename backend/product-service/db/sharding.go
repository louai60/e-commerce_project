package db

import (
	"context"
	"fmt"
	"hash/crc32"
	"strconv"
	"sync"
)

// ShardKey represents a key used for sharding
type ShardKey string

// ShardingStrategy defines the interface for determining which shard to use
type ShardingStrategy interface {
	// GetShardID returns the shard ID for a given shard key
	GetShardID(key ShardKey) int
	// GetShardCount returns the total number of shards
	GetShardCount() int
}

// ModuloShardingStrategy implements a simple modulo-based sharding strategy
type ModuloShardingStrategy struct {
	shardCount int
}

// NewModuloShardingStrategy creates a new modulo-based sharding strategy
func NewModuloShardingStrategy(shardCount int) *ModuloShardingStrategy {
	if shardCount <= 0 {
		shardCount = 1
	}
	return &ModuloShardingStrategy{
		shardCount: shardCount,
	}
}

// GetShardID returns the shard ID for a given shard key using modulo
func (s *ModuloShardingStrategy) GetShardID(key ShardKey) int {
	if s.shardCount <= 1 {
		return 0
	}

	// Use CRC32 to hash the key
	hash := crc32.ChecksumIEEE([]byte(key))
	return int(hash % uint32(s.shardCount))
}

// GetShardCount returns the total number of shards
func (s *ModuloShardingStrategy) GetShardCount() int {
	return s.shardCount
}

// ConsistentHashingStrategy implements consistent hashing for more stable sharding
type ConsistentHashingStrategy struct {
	virtualNodes int
	hashRing     []uint32
	nodeMap      map[uint32]int
	shardCount   int
	mu           sync.RWMutex
}

// NewConsistentHashingStrategy creates a new consistent hashing strategy
func NewConsistentHashingStrategy(shardCount, virtualNodes int) *ConsistentHashingStrategy {
	if shardCount <= 0 {
		shardCount = 1
	}
	if virtualNodes <= 0 {
		virtualNodes = 100
	}

	s := &ConsistentHashingStrategy{
		virtualNodes: virtualNodes,
		hashRing:     make([]uint32, 0, shardCount*virtualNodes),
		nodeMap:      make(map[uint32]int),
		shardCount:   shardCount,
	}

	s.initHashRing()
	return s
}

// initHashRing initializes the hash ring with virtual nodes
func (s *ConsistentHashingStrategy) initHashRing() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.hashRing = make([]uint32, 0, s.shardCount*s.virtualNodes)
	s.nodeMap = make(map[uint32]int)

	for shardID := 0; shardID < s.shardCount; shardID++ {
		for v := 0; v < s.virtualNodes; v++ {
			virtualNodeKey := fmt.Sprintf("shard-%d-vnode-%d", shardID, v)
			hash := crc32.ChecksumIEEE([]byte(virtualNodeKey))
			s.hashRing = append(s.hashRing, hash)
			s.nodeMap[hash] = shardID
		}
	}

	// Sort the hash ring
	sortUint32Slice(s.hashRing)
}

// GetShardID returns the shard ID for a given shard key using consistent hashing
func (s *ConsistentHashingStrategy) GetShardID(key ShardKey) int {
	if s.shardCount <= 1 {
		return 0
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.hashRing) == 0 {
		return 0
	}

	hash := crc32.ChecksumIEEE([]byte(key))

	// Binary search to find the first node with hash >= key's hash
	idx := searchUint32(s.hashRing, hash)
	if idx >= len(s.hashRing) {
		idx = 0 // Wrap around to the first node
	}

	return s.nodeMap[s.hashRing[idx]]
}

// GetShardCount returns the total number of shards
func (s *ConsistentHashingStrategy) GetShardCount() int {
	return s.shardCount
}

// Helper function to sort uint32 slice
func sortUint32Slice(slice []uint32) {
	// Simple bubble sort for demonstration
	// In production, use a more efficient sorting algorithm
	n := len(slice)
	for i := 0; i < n; i++ {
		for j := 0; j < n-i-1; j++ {
			if slice[j] > slice[j+1] {
				slice[j], slice[j+1] = slice[j+1], slice[j]
			}
		}
	}
}

// Helper function to search for a uint32 in a sorted slice
func searchUint32(slice []uint32, val uint32) int {
	// Simple binary search
	left, right := 0, len(slice)-1
	for left <= right {
		mid := (left + right) / 2
		if slice[mid] < val {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return left % len(slice)
}

// ShardManager manages database shards
type ShardManager struct {
	strategy    ShardingStrategy
	shardDBs    []*DBConfig
	defaultDB   *DBConfig
	initialized bool
	mu          sync.RWMutex
}

// NewShardManager creates a new shard manager
func NewShardManager(defaultDB *DBConfig, strategy ShardingStrategy) *ShardManager {
	return &ShardManager{
		strategy:    strategy,
		defaultDB:   defaultDB,
		shardDBs:    make([]*DBConfig, strategy.GetShardCount()),
		initialized: false,
	}
}

// Initialize initializes the shard manager with database connections
func (m *ShardManager) Initialize(ctx context.Context, shardConfigs []ShardConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return nil
	}

	shardCount := m.strategy.GetShardCount()
	if len(shardConfigs) < shardCount {
		return fmt.Errorf("not enough shard configurations: got %d, need %d", len(shardConfigs), shardCount)
	}

	for i := 0; i < shardCount; i++ {
		// In a real implementation, you would create a new DBConfig for each shard using shardConfigs[i]
		// For now, we'll just use the default DB for all shards
		m.shardDBs[i] = m.defaultDB
	}

	m.initialized = true
	return nil
}

// GetShardForKey returns the database shard for a given key
func (m *ShardManager) GetShardForKey(key ShardKey) *DBConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized || len(m.shardDBs) == 0 {
		return m.defaultDB
	}

	shardID := m.strategy.GetShardID(key)
	if shardID < 0 || shardID >= len(m.shardDBs) {
		return m.defaultDB
	}

	return m.shardDBs[shardID]
}

// GetAllShards returns all database shards
func (m *ShardManager) GetAllShards() []*DBConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized || len(m.shardDBs) == 0 {
		return []*DBConfig{m.defaultDB}
	}

	return m.shardDBs
}

// ShardConfig represents the configuration for a database shard
type ShardConfig struct {
	ShardID  int
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// GetShardKeyFromID generates a shard key from an ID
func GetShardKeyFromID(id string) ShardKey {
	return ShardKey(id)
}

// GetShardKeyFromCustomerID generates a shard key from a customer ID
func GetShardKeyFromCustomerID(customerID string) ShardKey {
	return ShardKey("customer:" + customerID)
}

// GetShardKeyFromTenantID generates a shard key from a tenant ID
func GetShardKeyFromTenantID(tenantID string) ShardKey {
	return ShardKey("tenant:" + tenantID)
}

// GetShardKeyFromTimestamp generates a shard key from a timestamp
// This is useful for time-based sharding
func GetShardKeyFromTimestamp(timestamp int64) ShardKey {
	// Shard by month for time-based data
	// Format: YYYYMM
	return ShardKey(strconv.FormatInt(timestamp/2592000, 10))
}
