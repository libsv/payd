/*
required tables:
keys            - to store all our extended private keys created
paymentOutputs  - to store the outputs generated in PaymentRequests
txos            - to store our outputs and note when they have been spent

 */
CREATE TABLE keys (
    name   VARCHAR NOT NULL PRIMARY KEY
    ,xprv     VARCHAR NOT NULL
    ,createdAt DATETIME(3) NOT NULL
);
-- TODO - we will maybe need a payments table as an invoice can have many payments
CREATE TABLE invoices (
    paymentID VARCHAR PRIMARY KEY
    ,satoshis INTEGER
    ,paymentReceivedAt  TIMESTAMP
);

CREATE TABLE script_keys(
    ID INTEGER PRIMARY KEY AUTO INCREMENT -- TODO we may not need this?
    ,lockingscript TEXT NOT NULL PRIMARY KEY
    ,keyname TEXT NOT NULL
    ,derivationPath TEXT NOT NULL
)

!-- store unspent transactions
CREATE TABLE txos (
    outpoint       VARCHAR NOT NULL PRIMARY KEY
    ,txid           CHAR(64) NOT NULL CHECK (LENGTH(txid) = 64)
    ,vout		    BIGINT NOT NULL CHECK (vout >= 0 AND vout < 4294967296)
    ,keyname		TEXT NOT NULL
    ,derivationpath TEXT NOT NULL
    ,lockingscript  TEXT NOT NULL
    ,satoshis       BIGINT NOT NULL CHECK (satoshis >= 0)
    ,spentat        INTEGER(4) -- this is the date when YOU use the funds
    ,spendingtxid   CHAR(64) CHECK (LENGTH(txid) = 64) -- the txid where you'd spent this output
    ,createdAt      DATETIME(3) NOT NULL
    ,modifiedAt     DATETIME(3) NOT NULL
 );

CREATE TABLE transactions (
    txid            CHAR(64) NOT NULL CHECK (LENGTH(txid) = 64) PRIMARY KEY
    ,paymentID VARCHAR
    ,txhex          TEXT NOT NULL
    ,createdAt      DATETIME(3) NOT NULL
)





