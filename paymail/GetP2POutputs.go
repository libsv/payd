package paymail

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mrz1836/paymail-inspector/paymail"
)

// GetP2POutputs will return list of outputs for the P2P transactions to use
func GetP2POutputs(paymailAddress string, sats uint64) (string, []*paymail.Output, error) {
	paymailInput := parseIfHandcashHandle(paymailAddress)
	domain, address := paymail.ExtractParts(paymailInput)

	// Did we get a paymail address?
	if len(address) == 0 {
		return "", nil, errors.New("paymail address not found or invalid")
	}

	// Validate the paymail address and domain (error already shown)
	if err := validatePaymailAndDomain(address, domain); err != nil {
		return "", nil, err
	}

	parts := strings.Split(address, "@")
	alias := parts[0]

	capability := GlobalPaymailCapabilities[domain]

	// Extract the URL from the capabilities response
	p2pURL := capability.GetString(paymail.BRFCP2PPaymentDestination, "")
	if len(p2pURL) == 0 {
		err := fmt.Errorf("the provider %s is missing a required capability: %s", domain, paymail.BRFCP2PPaymentDestination)
		return "", nil, err
	}

	p2pResponse, err := paymail.GetP2PPaymentDestination(p2pURL, alias, domain, &paymail.P2PPaymentDestinationRequest{Satoshis: sats}, true)
	if err != nil {
		return "", nil, err
	}

	// Test the status code
	if p2pResponse.StatusCode != http.StatusOK && p2pResponse.StatusCode != http.StatusNotModified {
		// Paymail address not found?
		if p2pResponse.StatusCode == http.StatusNotFound {
			err = fmt.Errorf("paymail address not found")
		} else {
			err = fmt.Errorf("bad response from paymail provider: %d", p2pResponse.StatusCode)
		}

		return "", nil, err
	}

	return p2pResponse.Reference, p2pResponse.Outputs, nil

}
