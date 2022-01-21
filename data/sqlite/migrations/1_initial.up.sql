-- required tables:
-- keys            - to store all our extended private keys created
-- paymentOutputs  - to store the outputs generated in PaymentRequests
-- txos            - to store our outputs and note when they have been spent

CREATE TABLE users(
    user_id         INTEGER PRIMARY KEY AUTOINCREMENT
    ,is_owner       BOOLEAN NOT NULL DEFAULT 0
    ,name           VARCHAR NOT NULL
    ,avatar_url     VARCHAR
    ,email          VARCHAR NOT NULL
    ,address        VARCHAR
    ,phone_number   VARCHAR
);

CREATE TABLE keys (
    name        VARCHAR NOT NULL
    ,user_id     INTEGER NOT NULL
    ,xprv       VARCHAR NOT NULL
    ,createdAt  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ,FOREIGN KEY (user_id) REFERENCES users(user_id)
);

CREATE TABLE users_meta(
    user_id         INTEGER NOT NULL
    ,key            VARCHAR NOT NULL
    ,value          VARCHAR NOT NULL
    ,FOREIGN KEY (user_id) REFERENCES users(user_id)
    ,CONSTRAINT users_key UNIQUE(user_id, key)
);

CREATE TABLE users_peerchannels(
    id                          INTEGER PRIMARY KEY AUTOINCREMENT
    ,user_id                    INTEGER
    ,account_id                 INTEGER NOT NULL
    ,user_name                  VARCHAR NOT NULL
    ,password                   VARCHAR NOT NULL
    ,FOREIGN KEY (user_id) REFERENCES users(user_id)
    ,CONSTRAINT user_id_key UNIQUE(user_id)
    ,CONSTRAINT account_id_key UNIQUE(account_id)
);

CREATE TABLE peerchannels(
    id                          INTEGER PRIMARY KEY AUTOINCREMENT
    ,peerchannels_account_id    INTEGER NOT NULL
    ,channel_id                 VARCHAR NOT NULL
    ,channel_host               VARCHAR NOT NULL
    ,channel_type               VARCHAR NOT NULL
    ,closed                     BOOLEAN NOT NULL DEFAULT 0
    ,FOREIGN KEY (peerchannels_account_id) REFERENCES users_peerchannels(account_id)
    ,CONSTRAINT channel_id_host_key UNIQUE(channel_id, channel_host)
);

CREATE TABLE peerchannels_api_tokens(
    id                          INTEGER PRIMARY KEY AUTOINCREMENT
    ,peerchannels_channel_id    VARCHAR NOT NULL
    ,token                      VARCHAR NOT NULL
    ,role                       VARCHAR NOT NULL
    ,can_read                   BOOLEAN NOT NULL
    ,can_write                  BOOLEAN NOT NULL
    --,FOREIGN KEY (peerchannels_channel_id) REFERENCES peerchannels(channel_id)
    ,CONSTRAINT token_key UNIQUE(token)
);


-- TODO - we will maybe need a payments table as an invoice can have many payments
CREATE TABLE invoices (
    invoice_id              VARCHAR PRIMARY KEY
    ,satoshis               INTEGER NOT NULL
    ,payment_reference      VARCHAR(32)
    ,description            VARCHAR(1024)
    ,spv_required           BOOLEAN NOT NULL DEFAULT 0
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

CREATE TABLE fee_rates (
	fee_rate_id     INTEGER PRIMARY KEY
	,invoice_id     VARCHAR
	,fee_json       TEXT
	,expires_at     TIMESTAMP
	,FOREIGN KEY (invoice_id) REFERENCES invoices(invoice_id)
    ,CONSTRAINT fee_key UNIQUE(invoice_id)
);

CREATE TABLE transactions (
    tx_id               CHAR(64) NOT NULL PRIMARY KEY
    ,tx_hex             TEXT NOT NULL
    ,state VARCHAR(10)  NOT NULL DEFAULT 'pending'
    ,fail_reason        TEXT
    ,created_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ,updated_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ,deleted_at         TIMESTAMP
);

CREATE TABLE transaction_invoice (
    tx_id               CHAR(64) NOT NULL
    ,invoice_id         VARCHAR NOT NULL
    ,FOREIGN KEY (invoice_id) REFERENCES invoices(invoice_id)
    ,FOREIGN KEY (tx_id) REFERENCES transactions(tx_id)
);

CREATE TABLE destinations(
    destination_id  INTEGER PRIMARY KEY AUTOINCREMENT
    ,locking_script  VARCHAR(50) NOT NULL
    ,satoshis        BIGINT NOT NULL
    ,derivation_path TEXT NOT NULL
    ,key_name        VARCHAR NOT NULL DEFAULT 'masterkey'
    ,user_id     INTEGER NOT NULL
    ,state           VARCHAR(10) NOT NULL
    ,created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ,updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ,deleted_at      TIMESTAMP
    ,FOREIGN KEY (user_id) REFERENCES users(user_id)
    ,CONSTRAINT destinations_locking_script UNIQUE(locking_script)
);

CREATE INDEX idx_destinations_locking_script ON invoices (payment_reference);
CREATE INDEX idx_destinations_derivation_path ON destinations (derivation_path);

CREATE TABLE destination_invoice(
    destination_id  INTEGER NOT NULL,
    invoice_id      VARCHAR NOT NULL,
    FOREIGN KEY (destination_id) REFERENCES destinations(destination_id),
    FOREIGN KEY (invoice_id) REFERENCES invoices(invoice_id)
);

-- store unspent transactions
CREATE TABLE txos (
    outpoint        VARCHAR PRIMARY KEY,
    destination_id  INTEGER,
    tx_id           CHAR(64),
    vout		    BIGINT,
    spent_at        TIMESTAMP, -- this is the date when YOU use the funds
    spending_txid   CHAR(64), -- the txid where you'd spent this output
    reserved_for    VARCHAR, -- the paymentId of this txo is being spent against
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tx_id) REFERENCES transactions(tx_id),
    FOREIGN KEY (spending_txid) REFERENCES transactions(tx_id),
    FOREIGN KEY (destination_id) REFERENCES destinations(destination_id)
 );

CREATE TABLE proofs(
    blockhash           VARCHAR(255) NOT NULL
    ,tx_id              VARCHAR(64) NOT NULL
    ,data               TEXT NOT NULL
    ,created_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ,updated_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    ,PRIMARY KEY(blockhash, tx_id)
    ,FOREIGN KEY (tx_id) REFERENCES transactions(tx_id)
);

CREATE TABLE proof_callbacks(
    invoice_id                          VARCHAR NOT NULL,
    url                                 VARCHAR NOT NULL,
    token                               VARCHAR,
    state                               VARCHAR NOT NULL,
    attempts                            INTEGER NOT NULL DEFAULT 0,
    created_at                          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at                          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (invoice_id)            REFERENCES invoices(invoice_id),
    PRIMARY KEY(invoice_id,url)
);

INSERT INTO users(user_id, name, is_owner, avatar_url, email, address, phone_number)
VALUES(0, 'Userless', 0, '', 'user@less.com', '123 Street Fake', '123456789'),
      (1, 'Epictetus', 1, 'https://thispersondoesnotexist.com/image', 'epic@nchain.com', '1 Athens Avenue', '0800-call-me');

INSERT INTO users_meta(user_id, key, value)
VALUES(1, 'likes', 'Stoicism & placeholder data'),
      (1, 'dislikes', 'Malfeasance');

INSERT INTO users_peerchannels(user_id, account_id, user_name, password)
VALUES(0, 0, '', ''), -- userless, for receiving change proofs
      (1, 1, 'username', 'password');
