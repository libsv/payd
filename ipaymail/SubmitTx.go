package ipaymail

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tonicpow/go-paymail"
)

// TransactionSubmitter is a stub interface for now, to decouple, bip270 and make it unit testable.
type TransactionSubmitter interface {
	SubmitTx(paymailAddress, txHexStr, reference string) (txid, note string, err error)
}

type transactionService struct {
}

func NewTransactionService() *transactionService {
	return &transactionService{}
}

// SubmitTx function
func (t *transactionService) SubmitTx(paymailAddress, txHexStr, reference string) (txid, note string, err error) {
	// Set the domain and paymail
	alias, domain, address := paymail.SanitizePaymail(paymail.ConvertHandle(paymailAddress, false))

	// Did we get a paymail address?
	if len(address) == 0 {
		return "", "", errors.New("paymail address not found or invalid")
	}

	// Validate the paymail address and domain (error already shown)
	if err := validatePaymailAndDomain(address, domain); err != nil {
		return "", "", err
	}

	capability := GlobalPaymailCapabilities[domain]

	// Extract the URL from the capabilities response
	p2pDestinationURL := capability.GetString(paymail.BRFCP2PPaymentDestination, "")
	if len(p2pDestinationURL) == 0 {
		err := fmt.Errorf("the provider %s is missing a required capability: %s", domain, paymail.BRFCP2PPaymentDestination)
		return "", "", err
	}

	p2pRequest := &paymail.P2PTransaction{
		Hex:       txHexStr,
		Reference: reference,
	}

	p2pRequest.MetaData = new(paymail.P2PMetaData)

	// Fire the tx to the P2P endpoint
	var p2pResponse *paymail.P2PTransactionResponse
	if p2pResponse, err = submitTx(p2pDestinationURL, alias, domain, satoshis, p2pRequest); err != nil {
		return "", "", err
	}

	// Test the status code
	if p2pResponse.StatusCode != http.StatusOK && p2pResponse.StatusCode != http.StatusNotModified {
		// Paymail address not found?
		if p2pResponse.StatusCode == http.StatusNotFound {
			err = fmt.Errorf("paymail address not found")
		} else {
			err = fmt.Errorf("bad response from paymail provider: %d", p2pResponse.StatusCode)
		}

		return "", "", err
	}

	return p2pResponse.TxID, p2pResponse.Note, nil
}
