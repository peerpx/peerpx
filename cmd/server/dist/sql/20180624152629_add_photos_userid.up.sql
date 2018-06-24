ALTER TABLE photos ADD user_id int NULL;
CREATE INDEX idx_photos_userid ON photos (user_id);