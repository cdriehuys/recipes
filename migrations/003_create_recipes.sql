CREATE TABLE recipes (
    id uuid PRIMARY KEY,
    owner text NOT NULL REFERENCES "users" (id)
        ON DELETE CASCADE,
    title text NOT NULL,
    instructions text NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

{{ template "shared/update_time.sql" "recipes" }}

---- create above / drop below ----

DROP TABLE recipes;
