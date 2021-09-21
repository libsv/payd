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
    invoice_id              VARCHAR PRIMARY KEY
    ,satoshis               INTEGER NOT NULL
    ,payment_reference      VARCHAR(32)
    ,description            VARCHAR(1024)
    ,expires_at             TIMESTAMP
    ,payment_received_at    TIMESTAMP
    ,refund_to              VARCHAR
    ,refunded_at            TIMESTAMP
    ,state                  VARCHAR(10)
    ,created_at             TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ,updated_at             TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ,deleted_at             TIMESTAMP
);

CREATE INDEX idx_invoices_payment_reference ON invoices (payment_reference);

CREATE TABLE transactions (
    txid            CHAR(64) NOT NULL PRIMARY KEY
    ,invoice_id      VARCHAR NOT NULL
    ,tx_hex          TEXT NOT NULL
    ,created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ,FOREIGN KEY (invoice_id) REFERENCES invoices(invoice_id)
);

CREATE TABLE destinations(
    destination_id INTEGER PRIMARY KEY AUTOINCREMENT,
    locking_script VARCHAR(50) NOT NULL,
    satoshis       BIGINT NOT NULL,
    derivation_path TEXT NOT NULL,
    key_name VARCHAR NOT NULL,
    state VARCHAR(10) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (key_name) REFERENCES keys(name),
    CONSTRAINT destinations_locking_script UNIQUE(locking_script)
);

CREATE INDEX idx_destinations_locking_script ON invoices (payment_reference);
CREATE INDEX idx_destinations_derivation_path ON destinations (derivation_path);

CREATE TABLE destination_invoice(
    destination_id INTEGER,
    invoice_id VARCHAR,
    FOREIGN KEY (destination_id) REFERENCES destinations(destination_id),
    FOREIGN KEY (invoice_id) REFERENCES invoices(invoice_id)
);

-- store unspent transactions
CREATE TABLE txos (
    outpoint        VARCHAR,
    destination_id INTEGER,
    txid           CHAR(64),
    vout		   BIGINT,
    spent_at        TIMESTAMP, -- this is the date when YOU use the funds
    spending_txid   CHAR(64), -- the txid where you'd spent this output
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modified_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (txid) REFERENCES transactions(txid),
    FOREIGN KEY (spending_txid) REFERENCES transactions(txid),
    FOREIGN KEY (destination_id) REFERENCES destinations(destination_id),
    CONSTRAINT outpoint_key UNIQUE(outpoint)
 );

CREATE TABLE proofs(
    blockhash VARCHAR(255) NOT NULL
    ,txid  VARCHAR(64) NOT NULL
    ,data TEXT NOT NULL
    ,createdAt      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ,modifiedAt     TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ,PRIMARY KEY(blockhash, txid)
    ,FOREIGN KEY (txid) REFERENCES transactions(txid)
);

INSERT INTO keys(name, xprv)
VALUES('masterkey','11111111111112xVQYuzHSiJmG55ahUXStc73UpffdMqgy4GTd4B5TXbn1ZY16Derh4uaoVyK4ZkCbn8GcDvV8GzLAcsDbdzUkgafnKPW6Nj');


