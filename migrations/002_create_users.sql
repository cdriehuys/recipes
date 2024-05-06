CREATE TABLE users (
    id text NOT NULL PRIMARY KEY,
    "name" text NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_login TIMESTAMPTZ
);

{{ template "shared/update_time.sql" "users" }}

---- create above / drop below ----

DROP TABLE "users";
