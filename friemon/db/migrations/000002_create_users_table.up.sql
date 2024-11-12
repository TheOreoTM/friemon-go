CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    balance INT NOT NULL DEFAULT 0,
    selected_id UUID,
    order_by INT NOT NULL DEFAULT 0,
    order_desc BOOLEAN NOT NULL DEFAULT FALSE,
    shinies_caught INT NOT NULL DEFAULT 0,
    next_idx INT NOT NULL DEFAULT 1
);
