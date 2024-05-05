CREATE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
NEW.updated_at = now();
RETURN NEW;
END;
$$ language 'plpgsql';

---- create above / drop below ----

DROP FUNCTION update_modified_column;
