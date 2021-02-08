CREATE TABLE keys (
  keyname   VARCHAR
  ,xprv     VARCHAR
);

-- CREATE TABLE txos (
--  created        TIMESTAMPTZ NOT NULL
-- ,modified       TIMESTAMPTZ NOT NULL
-- ,instance       INTEGER NOT NULL
-- ,txid           CHAR(64) NOT NULL CHECK (LENGTH(txid) = 64)
-- ,vout				    BIGINT NOT NULL CHECK (vout >= 0 AND vout < 4294967296)
-- ,alias					TEXT NOT NULL
-- ,derivationpath TEXT NOT NULL
-- ,scriptpubkey   TEXT NOT NULL
-- ,satoshis       BIGINT NOT NULL CHECK (satoshis >= 0)
-- ,reservedat  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
-- ,spentat        TIMESTAMPTZ
-- ,spendingtxid   CHAR(64) CHECK (LENGTH(txid) = 64)
-- ,outpoint       VARCHAR NOT NULL
-- ,PRIMARY KEY (outpoint)
-- );

-- CREATE TABLE invoices (

-- );