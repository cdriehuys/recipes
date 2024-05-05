CREATE TABLE recipes (
    id uuid PRIMARY KEY,
    title text NOT NULL,
    instructions text NOT NULL
);

---- create above / drop below ----

DROP TABLE recipes;
