package sqlite

const (
	sqlTransactionCreate = `
		INSERT INTO transactions(txid, paymentID, txhex)
		VALUES(:txid, :paymentID, :txhex)
	`

	sqlTxoCreate = `
		INSERT INTO txos(outpoint, txid, vout, keyname, derivationpath, lockingscript, satoshis)
		VALUES(:outpoint, :txid, :vout, :keyname, :derivationpath, :lockingscript, :satoshis)
	`

	sqlTransactionByID = `
	SELECT txid, paymentID, txhex, createdAt
	FROM transactions
	WHERE txid = :txid
	`

	sqlTxosByTxID = `
	SELECT outpoint, txid, vout, keyname, derivationpath, lockingscript, satoshis, 
				spentat, spendingtxid, createdAt, modifiedAt 
	FROM txos
	WHERE txid = :txid
	`
)
