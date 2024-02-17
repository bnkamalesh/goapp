CREATE TABLE IF NOT EXISTS user_notes (
    id UUID PRIMARY KEY,
    title TEXT,
    content TEXT,
    user_id UUID references users(id),
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TRIGGER tr_users_bu BEFORE UPDATE on user_notes
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();