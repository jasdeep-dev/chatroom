CREATE TABLE IF NOT EXISTS Messages (
    TimeStamp TIMESTAMP,
    Text TEXT,
    Name TEXT,
    FOREIGN KEY (Name) REFERENCES Users(Name)
);