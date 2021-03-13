/*
required tables:
keys            - to store all our extended private keys created
paymentOutputs  - to store the outputs generated in PaymentRequests
txos            - to store our outputs and note when they have been spent

 */
CREATE TABLE keys (
    name        VARCHAR NOT NULL PRIMARY KEY
    ,xprv       VARCHAR NOT NULL
    ,createdAt  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- TODO - we will maybe need a payments table as an invoice can have many payments
CREATE TABLE invoices (
    paymentID           VARCHAR PRIMARY KEY
    ,satoshis           INTEGER NOT NULL
    ,paymentReceivedAt  TIMESTAMP
    ,refundTo           VARCHAR
);

CREATE TABLE script_keys(
    derivationID INTEGER NOT NULL
    ,lockingscript  TEXT NOT NULL PRIMARY KEY
    ,keyname        TEXT NOT NULL
    ,FOREIGN KEY (keyname) REFERENCES keys(name)
    ,FOREIGN KEY (derivationID) REFERENCES derivation_paths(ID)
);

CREATE TABLE derivation_paths(
    ID INTEGER PRIMARY KEY AUTOINCREMENT
    ,paymentID INTEGER NOT NULL
    ,path TEXT NOT NULL
    ,prefix TEXT NOT NULL
    ,pathIndex INTEGER NOT NULL
    ,createdAt      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ,FOREIGN KEY (paymentID) REFERENCES invoices(paymentID)
);

CREATE TABLE transactions (
    txid            CHAR(64) NOT NULL PRIMARY KEY
    ,paymentID      VARCHAR NOT NULL
    ,txhex          TEXT NOT NULL
    ,createdAt      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ,FOREIGN KEY (paymentID) REFERENCES invoices(paymentID)
);

-- store unspent transactions
CREATE TABLE txos (
    outpoint        VARCHAR NOT NULL PRIMARY KEY
    ,txid           CHAR(64) NOT NULL CHECK (LENGTH(txid) = 64)
    ,vout		    BIGINT NOT NULL CHECK (vout >= 0 AND vout < 4294967296)
    ,keyname		TEXT NOT NULL
    ,derivationpath TEXT NOT NULL
    ,lockingscript  TEXT NOT NULL
    ,satoshis       BIGINT NOT NULL CHECK (satoshis >= 0)
    ,spentat        INTEGER(4) -- this is the date when YOU use the funds
    ,spendingtxid   CHAR(64) CHECK (LENGTH(txid) = 64) -- the txid where you'd spent this output
    ,createdAt      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ,modifiedAt     TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ,FOREIGN KEY (txid) REFERENCES transactions(txid)
 );


INSERT INTO keys(name, xprv)
VALUES('keyname','11111111111112xVQYuzHSiJmG55ahUXStc73UpffdMqgy4GTd4B5TXbn1ZY16Derh4uaoVyK4ZkCbn8GcDvV8GzLAcsDbdzUkgafnKPW6Nj');


