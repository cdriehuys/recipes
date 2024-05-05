CREATE TRIGGER update_modified_time BEFORE UPDATE ON "{{ . }}" FOR EACH ROW EXECUTE PROCEDURE update_modified_column();
