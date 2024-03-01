DROP TABLE IF EXISTS messages;

CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP,
    text TEXT,
    user_id INTEGER
);

-- Seed data for messages table
INSERT INTO messages (timestamp, text, user_id) VALUES
    ('2024-03-01 12:00:00', 'Hello, this is Alice!', 1),
    ('2024-03-01 12:05:00', 'Hi, Alice!', 2),
    ('2024-03-01 12:10:00', 'Hey, Bob!', 1);