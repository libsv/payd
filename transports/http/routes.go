package http

// Routes used in the http handlers.
const (
	// Receive payment endpoints.
	RouteV1Invoice     = "api/v1/invoices/:invoiceID"
	RouteV1Invoices    = "api/v1/invoices"
	RouteV1Destination = "api/v1/destinations/:invoiceID"
	RouteV1Payment     = "api/v1/payments/:invoiceID"
	RouteV1Proofs      = "api/v1/proofs/:txid"
	RouteV1Connect     = "api/v1/socket/connect/:invoiceID"
	RouteV1Transaction = "api/v1/transactions/:invoiceID"

	// User management.
	RouteV1Balance = "api/v1/balance"
	RouteV1Owner   = "api/v1/owner"

	// Sending payments.
	RouteV1Pay           = "api/v1/pay"
	RouteV1UnsignedOffTx = "api/v1/txs/unsignedoff"
	// TODO - fix this endpoint def.
	RouteV1Submit = "api/v1/submit"

	RouteV1Health = "api/v1/health"
)
