package sqlite

const (
	sqlTransactionCreate = `
		INSERT INTO transaction(txid, paymentID, txhex, createdAt)
		VALUES(:txid, :paymentID, :txhex, :createdAt)
	`

	sqlTxoCreate = `
		INSERT INTO txos(outpoint, txid, vout, keyname, derivationpath, lockingscript, satoshis,  createdat, modifiedat)
		VALUES(:outpoint, :txid, :vout, :keyname, :derivationPath, :lockingscript, :satoshis, :createdAt, :modifiedAt)
	`

	sqlTransactionByID = `
	SELECT txid, paymentID, txhex, createdAt
	FROM transactions
	WHERE txid = :txId
	`

	sqlTxosByTxID = `
	SELECT outpoint, txid, vout, alias, derivationpath, lockingscript, satoshis, 
				spentat, spendingtxid, createdat, modifiedat 
	FROM txos
	WHERE txid = :txId
	`
)
