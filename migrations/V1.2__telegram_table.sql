CREATE SCHEMA telegram;

CREATE TABLE telegram.users
(
    id             int PRIMARY KEY,
    chat_id        int,
    first_name varchar(20)
)