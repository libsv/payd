CREATE TABLE paymail_handles (
    user_id INTEGER NOT NULL
    ,handle VARCHAR NOT NULL
    ,FOREIGN KEY (user_id) REFERENCES users(user_id)
    ,UNIQUE(handle)
);

INSERT INTO paymail_handles(user_id, handle)
VALUES(1, 'epic');