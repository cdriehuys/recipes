CREATE TABLE categories (
    id uuid PRIMARY KEY,
    owner text NOT NULL REFERENCES "users" (id)
        ON DELETE CASCADE,
    "name" text NOT NULL
        CONSTRAINT categories_name_len CHECK (length("name") < 51),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT categories_unq_name UNIQUE ("owner", "name")
);

{{ template "shared/update_time.sql" "categories" }}

---- create above / drop below ----

DROP TABLE categories;
