DROP TRIGGER IF EXISTS trigger_ensure_unique_user_id ON users;
DROP FUNCTION IF EXISTS ensure_unique_user_id();
DROP FUNCTION IF EXISTS generate_random_id();
DROP TABLE IF EXISTS users;
DROP SEQUENCE IF EXISTS users_id_seq;
