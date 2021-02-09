package queries

const (
	InsertTransaction = `
	INSERT INTO transaction(txid, txhex, createdAt)
	VALUES(:txid, :txhex, :createdAt)
	`

	InsertTxo = `
	INSERT INTO txos(outpoint, instance, txid, vout, alias, derivationpath, scriptpubkey, satoshis, reservedat, spentat, spendingtxid, createdat, modifiedat)
	VALUES(:outpoint, :instance, :txid, :vout, :alias, :derivationPath, :scriptPubKey, :satoshis, :reservedAt, :spentAt, :spendingTxID, :createdAt, :modifiedAt)
	`

	TransactionByID = `
	SELECT txid, txhex, createdAt
	FROM transactions
	WHERE txid = :txID
	`

	TxosByTxID = `
	SELECT outpoint, instance, txid, vout, alias, derivationpath, scriptpubkey, satoshis, 
				reservedat, spentat, spendingtxid, createdat, modifiedat 
	FROM txos
	WHERE txid = :txID
	`
)
