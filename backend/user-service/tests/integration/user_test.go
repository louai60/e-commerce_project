package integration

import (
	// "context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	// "time"

	// "github.com/louai60/e-commerce_project/backend/user-service/models"
	// "github.com/louai60/e-commerce_project/backend/user-service/repository"
	// "github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/require"
	_ "github.com/lib/pq"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	// Set up test database connection using environment variables
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		password = "root"
	}

	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		dbName = "test_db"
	}

	// Construct DSN from environment variables
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbName)

	// Allow override with TEST_DATABASE_URL if set
	if envDSN := os.Getenv("TEST_DATABASE_URL"); envDSN != "" {
		dsn = envDSN
	}

	var err error
	testDB, err = sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	defer testDB.Close()

	// Run migrations
	err = setupTestDatabase(testDB)
	if err != nil {
		panic(err)
	}

	// Run tests
	code := m.Run()

	// Clean up
	cleanupTestDatabase(testDB)

	os.Exit(code)
}

func setupTestDatabase(db *sql.DB) error {
	// Create users table
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id VARCHAR(36) PRIMARY KEY,
            email VARCHAR(255) UNIQUE NOT NULL,
            username VARCHAR(255) UNIQUE NOT NULL,
            password VARCHAR(255) NOT NULL,
            first_name VARCHAR(255),
            last_name VARCHAR(255),
            role VARCHAR(50) NOT NULL,
            created_at TIMESTAMP NOT NULL,
            updated_at TIMESTAMP NOT NULL,
            is_active BOOLEAN DEFAULT true
        )
    `)
	return err
}

func cleanupTestDatabase(db *sql.DB) {
	db.Exec("DROP TABLE IF EXISTS users")
}

// func TestUserRepository_Integration(t *testing.T) {
//     // Use the same connection string as TestMain
//     host := os.Getenv("POSTGRES_HOST")
//     if host == "" {
//         host = "localhost"
//     }
//
//     port := os.Getenv("POSTGRES_PORT")
//     if port == "" {
//         port = "5432"
//     }
//
//     user := os.Getenv("POSTGRES_USER")
//     if user == "" {
//         user = "postgres"
//     }
//
//     password := os.Getenv("POSTGRES_PASSWORD")
//     if password == "" {
//         password = "root"
//     }
//
//     dbName := os.Getenv("POSTGRES_DB")
//     if dbName == "" {
//         dbName = "test_db"
//     }
//
//     // Construct DSN from environment variables
//     dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
//         user, password, host, port, dbName)

//     repo, err := repository.NewPostgresRepository(dsn)
//     require.NoError(t, err)

//     ctx := context.Background()

//     // Test CreateUser
//     user := &models.User{
//         ID:        "test-id",
//         Email:     "test@example.com",
//         Username:  "testuser",
//         Password:  "hashedpassword",
//         FirstName: "Test",
//         LastName:  "User",
//         Role:      "user",
//         CreatedAt: time.Now(),
//         UpdatedAt: time.Now(),
//         IsActive:  true,
//     }

//     err = repo.CreateUser(ctx, user)
//     assert.NoError(t, err)

//     // Test GetUser
//     retrieved, err := repo.GetUser(ctx, user.ID)
//     assert.NoError(t, err)
//     assert.Equal(t, user.Email, retrieved.Email)

//     // Clean up
//     err = repo.DeleteUser(ctx, user.ID)
//     assert.NoError(t, err)
// }
