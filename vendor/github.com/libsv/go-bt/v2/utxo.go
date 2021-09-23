package bt

import (
	"encoding/hex"
	"encoding/json"

	"github.com/libsv/go-bt/v2/bscript"
)

// UTXO an unspent transaction output, used for creating inputs
type UTXO struct {
	TxID          []byte
	Vout          uint32
	LockingScript *bscript.Script
	Satoshis      uint64
}

type utxoJSON struct {
	TxID         string  `json:"txid"`
	Vout         uint32  `json:"vout"`
	ScriptPubKey string  `json:"scriptPubKey"`
	Amount       float64 `json:"amount"`
	Satoshis     uint64  `json:"satoshis"`
}

func (u *UTXO) UnmarshalJSON(body []byte) error {
	var j utxoJSON
	if err := json.Unmarshal(body, &j); err != nil {
		return err
	}

	txID, err := hex.DecodeString(j.TxID)
	if err != nil {
		return err
	}

	ls, err := bscript.NewFromHexString(j.ScriptPubKey)
	if err != nil {
		return err
	}

	u.TxID = txID
	u.LockingScript = ls
	u.Vout = j.Vout
	if j.Satoshis > 0 {
		u.Satoshis = j.Satoshis
	} else {
		u.Satoshis = uint64(j.Amount * 100000000)
	}

	return nil
}

func (u *UTXO) MarshalJSON() ([]byte, error) {
	return json.Marshal(utxoJSON{
		TxID:         hex.EncodeToString(u.TxID),
		Amount:       float64(u.Satoshis) / 100000000,
		Satoshis:     u.Satoshis,
		Vout:         u.Vout,
		ScriptPubKey: u.LockingScript.String(),
	})
}
