CREATE TABLE contract_values (
                                 id SERIAL PRIMARY KEY,
                                 contract_key TEXT UNIQUE NOT NULL,
                                 contract_value TEXT NOT NULL,
                                 created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
