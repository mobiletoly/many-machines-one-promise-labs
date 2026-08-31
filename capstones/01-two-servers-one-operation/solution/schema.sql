CREATE TABLE orders (
    order_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    operation_id TEXT NOT NULL,
    wine TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    status TEXT NOT NULL CHECK (status = 'accepted'),
    CONSTRAINT one_canonical_order_per_operation UNIQUE (operation_id)
);
