package models

import (
	"github.com/louai60/e-commerce_project/backend/product-service/db"
)

// Shardable is an interface for models that can be sharded
type Shardable interface {
	// GetShardKey returns the shard key for the model
	GetShardKey() db.ShardKey
}

// ShardKeyProvider is a function that provides a shard key for a model
type ShardKeyProvider func(model interface{}) db.ShardKey

// DefaultShardKeyProvider provides a default implementation for getting shard keys
func DefaultShardKeyProvider(model interface{}) db.ShardKey {
	if shardable, ok := model.(Shardable); ok {
		return shardable.GetShardKey()
	}
	
	// Default implementations for common models
	switch m := model.(type) {
	case *Product:
		return db.GetShardKeyFromID(m.ID)
	case *Brand:
		return db.GetShardKeyFromID(m.ID)
	case *Category:
		return db.GetShardKeyFromID(m.ID)
	case *ProductVariant:
		return db.GetShardKeyFromID(m.ProductID) // Shard variants with their parent product
	default:
		// For unknown models, use empty shard key (will use default shard)
		return ""
	}
}

// GetShardKey implements the Shardable interface for Product
func (p *Product) GetShardKey() db.ShardKey {
	// If tenant ID is available, use that for sharding
	if p.TenantID != nil && *p.TenantID != "" {
		return db.GetShardKeyFromTenantID(*p.TenantID)
	}
	
	// Otherwise use product ID
	return db.GetShardKeyFromID(p.ID)
}

// GetShardKey implements the Shardable interface for Brand
func (b *Brand) GetShardKey() db.ShardKey {
	// If tenant ID is available, use that for sharding
	if b.TenantID != nil && *b.TenantID != "" {
		return db.GetShardKeyFromTenantID(*b.TenantID)
	}
	
	// Otherwise use brand ID
	return db.GetShardKeyFromID(b.ID)
}

// GetShardKey implements the Shardable interface for Category
func (c *Category) GetShardKey() db.ShardKey {
	// If tenant ID is available, use that for sharding
	if c.TenantID != nil && *c.TenantID != "" {
		return db.GetShardKeyFromTenantID(*c.TenantID)
	}
	
	// Otherwise use category ID
	return db.GetShardKeyFromID(c.ID)
}

// GetShardKey implements the Shardable interface for ProductVariant
func (v *ProductVariant) GetShardKey() db.ShardKey {
	// Variants are always sharded with their parent product
	return db.GetShardKeyFromID(v.ProductID)
}
