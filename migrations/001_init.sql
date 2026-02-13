-- 001_init.sql
CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL DEFAULT 'user'
);

CREATE TABLE IF NOT EXISTS countries (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  customs_rate NUMERIC(6,4) NOT NULL,
  delivery_cost NUMERIC(12,2) NOT NULL
);

CREATE TABLE IF NOT EXISTS cars (
  id BIGSERIAL PRIMARY KEY,
  brand TEXT NOT NULL,
  model TEXT NOT NULL,
  year INT NOT NULL,
  engine_volume INT NOT NULL,
  engine_type TEXT NOT NULL,
  base_price NUMERIC(12,2) NOT NULL,
  country_id BIGINT NOT NULL REFERENCES countries(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS calculations (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  car_id BIGINT NOT NULL REFERENCES cars(id) ON DELETE CASCADE,
  total_price NUMERIC(12,2) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS warnings (
  id BIGSERIAL PRIMARY KEY,
  calculation_id BIGINT NOT NULL REFERENCES calculations(id) ON DELETE CASCADE,
  message TEXT NOT NULL
);

INSERT INTO countries (name, customs_rate, delivery_cost)
VALUES
  ('Japan', 0.12, 800),
  ('Germany', 0.10, 900),
  ('USA', 0.14, 950)
ON CONFLICT DO NOTHING;

INSERT INTO cars (brand, model, year, engine_volume, engine_type, base_price, country_id)
VALUES
  ('Toyota', 'Camry', 2018, 2500, 'Petrol', 15000, 1),
  ('BMW', 'X5', 2016, 3000, 'Diesel', 28000, 2),
  ('Ford', 'Mustang', 2019, 5000, 'Petrol', 32000, 3)
ON CONFLICT DO NOTHING;
