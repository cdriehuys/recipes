ALTER TABLE recipes ADD COLUMN category uuid REFERENCES categories (id)
    ON DELETE SET NULL;

---- create above / drop below ----

ALTER TABLE recipes DROP COLUMN category;
