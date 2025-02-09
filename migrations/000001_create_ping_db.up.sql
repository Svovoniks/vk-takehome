CREATE SCHEMA IF NOT EXISTS main;

CREATE TABLE IF NOT EXISTS main.ping_event (
    ip TEXT UNIQUE,
    ping_ms FLOAT,
    pinged_at TIMESTAMP
);
