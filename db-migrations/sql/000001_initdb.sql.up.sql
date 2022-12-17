BEGIN;

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
CREATE INDEX immis_id_idx ON immis(id);

COMMIT;
