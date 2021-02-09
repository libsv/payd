CREATE TABLE keys (
    name   VARCHAR NOT NULL PRIMARY KEY
    ,xprv     VARCHAR NOT NULL
    ,createdAt DATETIME(3) NOT NULL
);

CREATE TABLE txos (
    outpoint       VARCHAR NOT NULL PRIMARY KEY
    ,instance       INTEGER NOT NULL
    ,txid           CHAR(64) NOT NULL CHECK (LENGTH(txid) = 64)
    ,vout		    BIGINT NOT NULL CHECK (vout >= 0 AND vout < 4294967296)
    ,alias			TEXT NOT NULL
    ,derivationpath TEXT NOT NULL
    ,scriptpubkey   TEXT NOT NULL
    ,satoshis       BIGINT NOT NULL CHECK (satoshis >= 0)
    ,reservedat     DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
    ,spentat        INTEGER(4)
    ,spendingtxid   CHAR(64) CHECK (LENGTH(txid) = 64)
    ,createdAt        DATETIME(3) NOT NULL
    ,modifiedAt       DATETIME(3) NOT NULL
 );

CREATE TABLE transactions (
    txid            CHAR(64) NOT NULL CHECK (LENGTH(txid) = 64) PRIMARY KEY
    ,txhex          TEXT NOT NULL
    ,createdAt        DATETIME(3) NOT NULL
)





