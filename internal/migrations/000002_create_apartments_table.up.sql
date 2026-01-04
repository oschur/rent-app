CREATE TABLE IF NOT EXISTS apartments (
    id SERIAL PRIMARY KEY,
    owner_id INT NOT NULL,
    status VARCHAR(64) NOT NULL,
    price_unit VARCHAR(64) NOT NULL,
    title TEXT,
    price INT NOT NULL,
    country VARCHAR(64) NOT NULL,
    city VARCHAR(64) NOT NULL,
    address VARCHAR(128) NOT NULL,
    area_m2 INT NOT NULL,
    rooms INT NOT NULL,
    floor INT,
    total_floors INT,
    pets_allowed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_apartments_owner FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_apartments_owner_id ON apartments(owner_id);
CREATE INDEX IF NOT EXISTS idx_apartments_status ON apartments(status);
CREATE INDEX IF NOT EXISTS idx_apartments_country ON apartments(country);
CREATE INDEX IF NOT EXISTS idx_apartments_city ON apartments(city);
CREATE INDEX IF NOT EXISTS idx_apartments_price ON apartments(price);
CREATE INDEX IF NOT EXISTS idx_apartments_rooms ON apartments(rooms);
CREATE INDEX IF NOT EXISTS idx_apartments_city_price ON apartments(city, price);
CREATE INDEX IF NOT EXISTS idx_apartments_country_city ON apartments(country, city);