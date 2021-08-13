package http

// Routes used in the http handlers.
const (
	RoutePaymentRequest = "r/:paymentID"
	RoutePayment        = "payment/:paymentID"

	RouteInvoice  = "api/v1/invoices/:paymentID"
	RouteInvoices = "api/v1/invoices"
	RouteBalance  = "api/v1/balance"

	RouteProofs = "api/v1/proofs/:txid"
)
