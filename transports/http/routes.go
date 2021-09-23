package http

// Routes used in the http handlers.
const (
	RouteV1Invoice  = "api/v1/invoices/:invoiceID"
	RouteV1Invoices = "api/v1/invoices"
	RouteV1Balance  = "api/v1/balance"
	RouteV1Owner    = "api/v1/owner"

	RouteV1Proofs      = "api/v1/proofs/:txid"
	RouteV1Destination = "api/v1/destinations/:invoiceID"
)
