CREATE TABLE IF NOT EXISTS updates (
    id        serial  PRIMARY KEY,
    room_name varchar NOT NULL,
    client_id bigint  NOT NULL,
    clock     bigint  NOT NULL,
    content   bytea   NOT NULL

--     PRIMARY KEY (room_name, client_id, clock)
);

-- CREATE INDEX IF NOT EXISTS updates_vector_clock on updates (client_id, clock);

CREATE TABLE IF NOT EXISTS removals (
    id        serial PRIMARY KEY,
    room_name varchar NOT NULL,
    content   bytea   NOT NULL
);