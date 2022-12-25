BEGIN;

-- TODO: CREATE INDEX for all the tables

CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  username TEXT NOT NULL,
  email_address TEXT NOT NULL,
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
ALTER TABLE immis ADD CONSTRAINT immis_fk_users
  FOREIGN KEY (user_id) REFERENCES users(id);

---

CREATE TABLE listys (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  route_name TEXT NOT NULL,
  display_name TEXT NOT NULL,
  ctime TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

ALTER TABLE listys ADD CONSTRAINT listys_fk_accounts
  FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE listys ADD CONSTRAINT listys_unique_user_id__route_name
  UNIQUE (user_id, route_name);
ALTER TABLE listys ADD CONSTRAINT listys_unique_user_id__display_name
  UNIQUE (user_id, display_name);

---

CREATE TABLE graf (
  listy_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  ctime TIMESTAMP WITHOUT TIME ZONE NOT NULL
);
ALTER TABLE graf ADD CONSTRAINT graf_fk_accounts
  FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE graf ADD CONSTRAINT graf_fk_listys
  FOREIGN KEY (listy_id) REFERENCES listys(id);
ALTER TABLE graf ADD CONSTRAINT graf_unique_listy_id__user_id
  UNIQUE (user_id, listy_id);

COMMIT;
