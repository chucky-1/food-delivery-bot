CREATE SCHEMA internal;

CREATE TABLE internal.users
(
    id              uuid PRIMARY KEY,
    telegram_id     int,
    organization_id uuid
);

CREATE TABLE internal.organizations
(
    id       uuid PRIMARY KEY,
    name     varchar(100),
    address  varchar(150),
    owner_id uuid,
    lunch_time interval
);

CREATE TABLE internal.orders
(
    date             date,
    user_telegram_id int,
    dish_name        varchar(100),
    dish_price       float,
    category         varchar(20),
    confirmed        boolean DEFAULT false
);

CREATE TABLE internal.statistics
(
    date            date,
    organization_id uuid,
    order_amount    real
)