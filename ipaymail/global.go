package ipaymail

import "github.com/tonicpow/go-paymail"

// GlobalPaymailCapabilities var
var GlobalPaymailCapabilities = make(map[string]*paymail.Capabilities) // TODO: use redis

// ReferencesMap var
var ReferencesMap = make(map[string]string) // TODO: use redis and put reference in invoice object

// PaymailInit func
func PaymailInit() {

	// Custom name server for DNS resolution (looking for the SRV record)
	nameServer = defaultNameServer

	// Custom service name for the SRV record
	serviceName = paymail.DefaultServiceName

	// Custom protocol for the SRV record
	protocol = paymail.DefaultProtocol

	// Custom port for the SRV record (target address)
	port = paymail.DefaultPort

	// Custom priority for the SRV record
	priority = paymail.DefaultPriority

	// Custom weight for the SRV record
	weight = paymail.DefaultWeight

	// Run the SRV check on the domain
	skipSrvCheck = false

	// Run the DNSSEC check on the target domain
	skipDNSCheck = false

	// Run the SSL check on the target domain
	skipSSLCheck = false
}
