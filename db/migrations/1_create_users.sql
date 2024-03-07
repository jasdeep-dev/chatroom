DROP TABLE IF EXISTS users;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE,
    is_online BOOLEAN,
    theme TEXT,
    preferred_username TEXT,
    given_name TEXT,
    family_name TEXT,
    email TEXT UNIQUE
);

-- Seed data for users table
INSERT INTO users (name, is_online, theme, preferred_username, given_name, family_name, email)
VALUES 
    ('John Doe', true, 'light', 'john_doe', 'John', 'Doe', 'john@example.com'),
    ('Jane Smith', false, 'dark', 'jane_smith', 'Jane', 'Smith', 'jane@example.com');
