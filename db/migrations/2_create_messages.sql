DROP TABLE IF EXISTS messages;

CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP,
    text TEXT,
    user_id varchar NOT NULL,
    group_id varchar,
    first_name varchar,
    email varchar
);

-- -- Seed data for messages table
-- INSERT INTO messages (timestamp, text, user_id, first_name, email) VALUES
--     ('2024-03-01 12:00:00', 'Hello, this is Alice!', '2f1ee47f-4201-49b8-8a8a-7dae2ec11e40', 'smiley', 'smiley@gmail.com'),
--     ('2024-03-01 12:05:00', 'Hi, Alice!', '2f1ee47f-4201-49b8-8a8a-7dae2ec11e40', 'smiley', 'smiley@gmail.com'),
--     ('2024-03-01 12:10:00', 'Hey, Bob!', '2f1ee47f-4201-49b8-8a8a-7dae2ec11e40', 'smiley', 'smiley@gmail.com');
-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.
