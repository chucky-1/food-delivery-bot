ALTER TABLE internal.users
    ALTER COLUMN telegram_id TYPE bigint;

ALTER TABLE internal.orders
    ALTER COLUMN user_telegram_id TYPE bigint;

ALTER TABLE telegram.users
    ALTER COLUMN id TYPE bigint,
    ALTER COLUMN chat_id TYPE bigint;