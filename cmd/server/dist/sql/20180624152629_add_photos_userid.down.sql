ALTER TABLE photos DROP user_id int;
DROP INDEX idx_photos_userid ON photos;