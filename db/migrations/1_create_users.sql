DROP TABLE IF EXISTS users;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE,
    is_online BOOLEAN,
    theme TEXT,
    password_hash TEXT
);

-- Seed data for users table
INSERT INTO users (name, is_online, theme, password_hash) VALUES
    ('Alice', true, 'default', '$2a$10$RlT/sPiel/dFy8jHlDxfT.XLRdTS3yM2tL4eugh7cJvS5B468tqlu'),
    ('Bob', false, 'dark', '$2a$10$RlT/sPiel/dFy8jHlDxfT.XLRdTS3yM2tL4eugh7cJvS5B468tqlu'),
    ('Charlie', true, 'light', '$2a$10$RlT/sPiel/dFy8jHlDxfT.XLRdTS3yM2tL4eugh7cJvS5B468tqlu');
