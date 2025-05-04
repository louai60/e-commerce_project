package service

// This file ensures that all necessary imports are properly included
// to avoid linting errors related to undefined variables.

import (
	// Import Cloudinary package to ensure it's available
	_ "github.com/cloudinary/cloudinary-go/v2"
	_ "github.com/cloudinary/cloudinary-go/v2/api/uploader"
)
