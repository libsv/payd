package paymail

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mrz1836/go-validate"
	"github.com/mrz1836/paymail-inspector/paymail"
)

var (
	defaultNameServer = "8.8.8.8"
)

// CapabilitySetup returns the CapabilitiesResponse for a specific domain (eg. handcash.io or moneybutton.com).
func CapabilitySetup(domain string) (capabilities *paymail.CapabilitiesResponse, err error) {
	capabilityDomain := ""
	capabilityPort := paymail.DefaultPort

	// Get the record
	if srv, err := paymail.GetSRVRecord(paymail.DefaultServiceName, paymail.DefaultProtocol, domain, defaultNameServer); err != nil {
		capabilityDomain = domain
	} else if srv != nil {
		capabilityDomain = srv.Target
		capabilityPort = int(srv.Port)
	}

	// Look up the capabilities
	capabilities, err = paymail.GetCapabilities(capabilityDomain, capabilityPort, true)

	return
}

// validatePaymailAndDomain will do a basic validation on the paymail format
func validatePaymailAndDomain(paymailAddress, domain string) error {

	// Validate the format for the paymail address (paymail addresses follow conventional email requirements)
	if ok, err := validate.IsValidEmail(paymailAddress, false); err != nil {
		return fmt.Errorf("paymail address failed format validation: %s", err.Error())
	} else if !ok {
		return fmt.Errorf("paymail address failed format validation: unknown reason")
	}

	// Check for a real domain (require at least one period)
	if !strings.Contains(domain, ".") {
		return fmt.Errorf("domain name is invalid: %s", domain)
	} else if !validate.IsValidDNSName(domain) { // Basic DNS check (not a REAL domain name check)
		return fmt.Errorf("domain name failed DNS check: %s", domain)
	}

	return nil
}

func parseIfHandcashHandle(paymail string) string {
	var validID = regexp.MustCompile(`^\$[a-zA-Z0-9\-_.]{4,}$`)

	if validID.MatchString(paymail) {
		return paymail[1:] + "@handcash.io"
	} else {
		return paymail
	}
}
