DROP TABLE IF EXISTS Organizations;

-- CREATE TABLE IF NOT EXISTS users (
--     id SERIAL PRIMARY KEY,
--     name TEXT UNIQUE,
--     is_online BOOLEAN,
--     theme TEXT,
--     preferred_username TEXT,
--     given_name TEXT,
--     family_name TEXT,
--     email TEXT UNIQUE
-- );

-- -- Seed data for users table
-- INSERT INTO users (name, is_online, theme, preferred_username, given_name, family_name, email)
-- VALUES 
--     ('John Doe', false, 'light', 'john_doe', 'John', 'Doe', 'john@example.com'),
--     ('Jane Smith', false, 'dark', 'jane_smith', 'Jane', 'Smith', 'jane@example.com');

-- organization Table
-- CREATE TABLE IF NOT EXISTS Organizations (
--     OrgID SERIAL PRIMARY KEY,
--     OrgName VARCHAR(100) NOT NULL,
--     Description TEXT,
--     CreatorID VARCHAR(50) REFERENCES Users(UserID),
--     CreationDate TIMESTAMP DEFAULT CURRENT_TIMESTAMP
--     -- Add other fields as needed
-- );

-- -- OrganizationMembership Table
-- CREATE TABLE IF NOT EXISTS OrganizationMemberships (
--     OrgMembershipID SERIAL PRIMARY KEY,
--     UserID VARCHAR(50) REFERENCES Users(UserID),
--     OrgID INT REFERENCES Organizations(OrgID),
--     Role VARCHAR(20) NOT NULL,
--     JoinDate TIMESTAMP DEFAULT CURRENT_TIMESTAMP
--     -- Add other fields as needed
-- );