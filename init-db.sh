#!/bin/bash
set -e

# Function to create a database if it doesn't exist
create_database() {
  local db_name=$1
  echo "Creating database: $db_name"
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    SELECT 'CREATE DATABASE $db_name' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$db_name')\gexec
    GRANT ALL PRIVILEGES ON DATABASE $db_name TO $POSTGRES_USER;
EOSQL
}

# Create all required databases
create_database "nexcart_product"
create_database "nexcart_user"
create_database "nexcart_inventory"
create_database "nexcart_order"

echo "All databases created successfully"
