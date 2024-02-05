ALTER TABLE telegram.users
    DROP COLUMN first_name,
    DROP COLUMN last_name;

ALTER TABLE internal.users
    ADD COLUMN first_name  varchar,
    ADD COLUMN last_name   varchar,
    ADD COLUMN middle_name varchar;