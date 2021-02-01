package ipaymail

import "github.com/tonicpow/go-paymail"

// GlobalPaymailCapabilities var
var GlobalPaymailCapabilities = make(map[string]*paymail.Capabilities) // TODO: use badger

// ReferencesMap var
var ReferencesMap = make(map[string]string) // TODO: use badger and put reference in invoice object
