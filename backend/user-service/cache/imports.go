package cache

// This file ensures that all necessary imports are properly included
// to avoid linting errors related to undefined variables.

import (
	// Import Redis package to ensure it's available
	_ "github.com/go-redis/redis/v8"
)
