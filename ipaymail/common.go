package ipaymail

// import (
// 	"fmt"
// 	"strings"

// 	"github.com/mrz1836/go-validate"
// 	"github.com/tonicpow/go-paymail"
// )

// // Creates a new client for Paymail
// func newPaymailClient() (*paymail.Client, error) {
// 	options, err := paymail.DefaultClientOptions()
// 	if err != nil {
// 		return nil, err
// 	}
// 	options.UserAgent = applicationFullName + ": v" + Version

// 	return paymail.NewClient(options, nil, nil)
// }

// // validatePaymailAndDomain will do a basic validation on the paymail format
// func validatePaymailAndDomain(paymailAddress, domain string) error {

// 	// Validate the format for the paymail address (paymail addresses follow conventional email requirements)
// 	if ok, err := validate.IsValidEmail(paymailAddress, false); err != nil {
// 		return fmt.Errorf("paymail address failed format validation: %s", err.Error())
// 	} else if !ok {
// 		return fmt.Errorf("paymail address failed format validation: unknown reason")
// 	}

// 	// Check for a real domain (require at least one period)
// 	if !strings.Contains(domain, ".") {
// 		return fmt.Errorf("domain name is invalid: %s", domain)
// 	} else if !validate.IsValidDNSName(domain) { // Basic DNS check (not a REAL domain name check)
// 		return fmt.Errorf("domain name failed DNS check: %s", domain)
// 	}

// 	return nil
// }
