CREATE TABLE IF NOT EXISTS products (
    id bigserial PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    image_url TEXT NOT NULL,
    average_rating REAL DEFAULT 0,
    version integer NOT NULL DEFAULT 1
);