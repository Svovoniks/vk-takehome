CREATE DATABASE ping_db;

\connect ping_db

CREATE SCHEMA main;

CREATE TABLE "main.ping_event" (
    ip TEXT UNIQUE,
    ping_ms FLOAT,
    pinged_at TIMESTAMP
);

