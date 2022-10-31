-- Keyspace creation.
CREATE KEYSPACE IF NOT EXISTS mcq_platform WITH replication = {'class' : 'SimpleStrategy', 'replication_factor' : 3};

-- User table creation.
CREATE TABLE IF NOT EXISTS mcq_platform.users (
    account_id  text,
    username    text,
    password    text,
    first_name  text,
    last_name   text,
    email       text,
    is_deleted  boolean,
    PRIMARY KEY ( (username, account_id) )
);
