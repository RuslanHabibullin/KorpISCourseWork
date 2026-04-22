-- +goose Up
-- +goose StatementBegin

-- Индекс для быстрого поиска заказов по клиенту
CREATE INDEX IF NOT EXISTS idx_work_order_client_id ON work_order(client_id);

-- Индекс для быстрого поиска заказов по статусу
CREATE INDEX IF NOT EXISTS idx_work_order_status ON work_order(status);

-- Индекс для быстрого поиска автомобилей по клиенту
CREATE INDEX IF NOT EXISTS idx_vehicle_client_id ON vehicle(client_id);

-- Индекс для быстрого поиска платежей по заказу
CREATE INDEX IF NOT EXISTS idx_payment_order_id ON payment(order_id);

-- Частичный индекс: открытые заказы (не закрытые)
CREATE INDEX IF NOT EXISTS idx_work_order_open
    ON work_order(created_at DESC)
    WHERE status <> 'closed';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_work_order_open;
DROP INDEX IF EXISTS idx_payment_order_id;
DROP INDEX IF EXISTS idx_vehicle_client_id;
DROP INDEX IF EXISTS idx_work_order_status;
DROP INDEX IF EXISTS idx_work_order_client_id;
-- +goose StatementEnd