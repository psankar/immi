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
  ctime TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  dbmtime TIMESTAMP WITHOUT TIME ZONE DEFAULT (now() AT TIME ZONE 'utc')
);

ALTER TABLE immis ADD CONSTRAINT immis_unique_id UNIQUE (id);
ALTER TABLE immis ADD CONSTRAINT immis_fk_users
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

---

CREATE TABLE listys (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  route_name TEXT NOT NULL,
  display_name TEXT NOT NULL,
  ctime TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  last_refresh_time TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

ALTER TABLE listys ADD CONSTRAINT listys_fk_accounts
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
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
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE graf ADD CONSTRAINT graf_fk_listys
  FOREIGN KEY (listy_id) REFERENCES listys(id) ON DELETE CASCADE;
ALTER TABLE graf ADD CONSTRAINT graf_unique_listy_id__user_id
  UNIQUE (user_id, listy_id);

---

CREATE TABLE tl(
  listy_id BIGINT NOT NULL,
  immi_id TEXT NOT NULL,
  dbctime TIMESTAMP WITHOUT TIME ZONE DEFAULT (now() AT TIME ZONE 'utc')
)

ALTER TABLE tl ADD CONSTRAINT tl_fk_listys
  FOREIGN KEY (listy_id) REFERENCES listys(id) ON DELETE CASCADE;
ALTER TABLE tl ADD CONSTRAINT tl_fk_immis
  FOREIGN KEY (immi_id) REFERENCES immis(id) ON DELETE CASCADE;

---

COMMIT;

-- TRUNCATE users CASCADE;
-- ALTER SEQUENCE users_id_seq RESTART WITH 1;
-- ALTER SEQUENCE listys_id_seq RESTART WITH 1;
