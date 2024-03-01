CREATE TABLE IF NOT EXISTS Users (
    Name TEXT PRIMARY KEY,
    IsOnline BOOLEAN,
    Theme TEXT,
    PasswordHash TEXT
);