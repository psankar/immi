BEGIN;

-- TODO: CREATE INDEX for all the tables

CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  username TEXT NOT NULL,
  email_address TEXT NOT NULL, -- TODO: Should we allow duplicates here ?
  password_hash TEXT NOT NULL,
  user_state TEXT NOT NULL -- Perhaps we can use enum here ?!
);

ALTER TABLE users ADD CONSTRAINT users_unique_username UNIQUE (username);

---

CREATE TABLE immis (
  -- primary key
  id TEXT NOT NULL,

  -- int of 4 bytes may be sufficient here and could provide better cache
  -- alignment by packing more records. But thinking positively for billions
  -- of users.
  user_id BIGINT NOT NULL,

  msg TEXT NOT NULL,
  ctime TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

ALTER TABLE immis ADD CONSTRAINT immis_unique_id UNIQUE (id);
ALTER TABLE immis ADD CONSTRAINT immis_fk_accounts
  FOREIGN KEY (user_id) REFERENCES users(id);

COMMIT;
