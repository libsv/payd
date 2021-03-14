package sqlite

const (
	sqlTransactionCreate = `
		INSERT INTO transaction(txid, paymentID, txhex)
		VALUES(:txid, :paymentID, :txhex)
	`

	sqlTxoCreate = `
		INSERT INTO txos(outpoint, txid, vout, keyname, derivationpath, lockingscript, satoshis)
		VALUES(:outpoint, :txid, :vout, :keyname, :derivationPath, :lockingscript, :satoshis)
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
