-- One-time PostgreSQL setup for the Debezium connector in
-- config/debezium/postgres-connector.json. Idempotent; run after the API has
-- started once (AutoMigrate creates the tables):
--
--   docker exec -i order-postgres psql -U postgres -d order_processing < config/debezium/setup.sql
--
-- cdc_user/cdc_password are local development credentials matching the
-- connector config. Change both for any shared environment.

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'cdc_user') THEN
        CREATE ROLE cdc_user WITH LOGIN REPLICATION PASSWORD 'cdc_password';
    END IF;
END
$$;

GRANT CONNECT ON DATABASE order_processing TO cdc_user;
GRANT USAGE ON SCHEMA public TO cdc_user;
GRANT SELECT ON users, categories, products, inventories, orders, order_items, payments, payment_webhooks TO cdc_user;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_publication WHERE pubname = 'order_processing_cdc') THEN
        CREATE PUBLICATION order_processing_cdc FOR TABLE
            users, categories, products, inventories, orders, order_items, payments, payment_webhooks;
    END IF;
END
$$;
