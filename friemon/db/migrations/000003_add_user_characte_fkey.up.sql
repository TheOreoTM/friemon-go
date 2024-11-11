ALTER TABLE users
  ADD CONSTRAINT selected_char_user_fk FOREIGN KEY (selected_id)
  REFERENCES characters (id)
  ON DELETE SET NULL;
