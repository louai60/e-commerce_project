package cache

// This file ensures that all necessary imports are properly included
// to avoid linting errors related to undefined variables.

import (
	// Import Redis package to ensure it's available
	_ "github.com/go-redis/redis/v8"
	// Import shared package to ensure it's available
	// _ "github.com/louai60/e-commerce_project/backend/shared"
)
