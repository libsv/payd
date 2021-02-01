package ipaymail

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mrz1836/paymail-inspector/paymail"
)

// SubmitTx function
func SubmitTx(paymailAddress, txHexStr, reference string) (string, string, error) {
	paymailInput := parseIfHandcashHandle(paymailAddress)
	domain, address := paymail.ExtractParts(paymailInput)

	// Did we get a paymail address?
	if len(address) == 0 {
		return "", "", errors.New("paymail address not found or invalid")
	}

	// Validate the paymail address and domain (error already shown)
	if err := validatePaymailAndDomain(address, domain); err != nil {
		return "", "", err
	}

	parts := strings.Split(address, "@")
	alias := parts[0]

	capability := GlobalPaymailCapabilities[domain]

	// Extract the URL from the capabilities response
	p2pDestinationURL := capability.GetString(paymail.BRFCP2PPaymentDestination, "")
	if len(p2pDestinationURL) == 0 {
		err := fmt.Errorf("the provider %s is missing a required capability: %s", domain, paymail.BRFCP2PPaymentDestination)
		return "", "", err
	}

	p2pRequest := &paymail.P2PTransactionRequest{
		Hex:       txHexStr,
		Reference: reference,
	}

	p2pRequest.MetaData = new(paymail.MetaData)

	p2pResponse, err := paymail.SendP2PTransaction(p2pDestinationURL, alias, domain, p2pRequest, true)
	if err != nil {
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
