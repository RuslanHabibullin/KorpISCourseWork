-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS client (
    client_id  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name  VARCHAR(255) NOT NULL,
    phone      VARCHAR(50)  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS vehicle (
    vehicle_id UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id  UUID         NOT NULL REFERENCES client(client_id) ON DELETE CASCADE,
    brand      VARCHAR(100) NOT NULL,
    model      VARCHAR(100) NOT NULL,
    plate      VARCHAR(20)  NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS work_order (
    order_id     UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id   UUID           NOT NULL REFERENCES vehicle(vehicle_id),
    client_id    UUID           NOT NULL REFERENCES client(client_id),
    status       VARCHAR(20)    NOT NULL DEFAULT 'draft'
                                CHECK (status IN ('draft','approved','in_progress','done','closed')),
    complaint    TEXT           NOT NULL DEFAULT '',
    total_amount NUMERIC(12,2)  NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS service (
    service_id UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255)   NOT NULL,
    base_price NUMERIC(12,2)  NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS work_order_service (
    id         UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id   UUID          NOT NULL REFERENCES work_order(order_id) ON DELETE CASCADE,
    service_id UUID          NOT NULL REFERENCES service(service_id),
    price      NUMERIC(12,2) NOT NULL,
    quantity   INT           NOT NULL DEFAULT 1 CHECK (quantity > 0)
);

CREATE TABLE IF NOT EXISTS part (
    part_id UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    name    VARCHAR(255)  NOT NULL,
    price   NUMERIC(12,2) NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS work_order_part (
    id       UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID          NOT NULL REFERENCES work_order(order_id) ON DELETE CASCADE,
    part_id  UUID          NOT NULL REFERENCES part(part_id),
    quantity INT           NOT NULL DEFAULT 1 CHECK (quantity > 0),
    price    NUMERIC(12,2) NOT NULL
);

CREATE TABLE IF NOT EXISTS stock (
    part_id UUID NOT NULL PRIMARY KEY REFERENCES part(part_id) ON DELETE CASCADE,
    qty     INT  NOT NULL DEFAULT 0 CHECK (qty >= 0)
);

CREATE TABLE IF NOT EXISTS payment (
    payment_id UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id   UUID          NOT NULL REFERENCES work_order(order_id) ON DELETE CASCADE,
    amount     NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    paid_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS payment;
DROP TABLE IF EXISTS stock;
DROP TABLE IF EXISTS work_order_part;
DROP TABLE IF EXISTS part;
DROP TABLE IF EXISTS work_order_service;
DROP TABLE IF EXISTS service;
DROP TABLE IF EXISTS work_order;
DROP TABLE IF EXISTS vehicle;
DROP TABLE IF EXISTS client;
-- +goose StatementEnd