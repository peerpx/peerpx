ALTER TABLE users ADD authuuid char(40) NULL;
CREATE UNIQUE INDEX users_authuuid_uindex ON users (authuuid);