package http

// Routes used in the http handlers.
const (
	RouteV1Invoice  = "api/v1/invoices/:invoiceID"
	RouteV1Invoices = "api/v1/invoices"
	RouteV1Balance  = "api/v1/balance"
	RouteV1Owner    = "api/v1/owner"
	RouteV1Pay      = "api/v1/pay"

	RouteV1Proofs      = "api/v1/proofs/:txid"
	RouteV1Destination = "api/v1/destinations/:invoiceID"
	RouteV1Payment     = "api/v1/payments/:invoiceID"

	RouteV1Connect = "api/v1/socket/connect/:invoiceID"
)
