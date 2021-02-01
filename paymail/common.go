package ipaymail

import "github.com/tonicpow/go-paymail"

// Creates a new client for Paymail
func newPaymailClient() (*paymail.Client, error) {
	options, err := paymail.DefaultClientOptions()
	if err != nil {
		return nil, err
	}
	options.UserAgent = applicationFullName + ": v" + Version

	return paymail.NewClient(options, nil, nil)
}
