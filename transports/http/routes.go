package http

// Routes used in the http handlers.
const (
	RoutePaymentRequest = "api/v1/payment/:paymentID"
	RoutePayment        = "api/v1/payment/:paymentID"

	RouteInvoice  = "api/v1/invoices/:paymentID"
	RouteInvoices = "api/v1/invoices"
	RouteBalance  = "api/v1/balance"

	RouteProofs   = "api/v1/proofs/:txid"
	RouteTxStatus = "api/v1/txstatus/:txid"

	RouteFundAdd           = "api/v1/funds"
	RouteFundGet           = "api/v1/funds"
	RouteFundSpend         = "api/v1/funds/spend"
	RouteFundRequestAmount = "api/v1/funds/:amount"

	RouteFundAndSign = "api/v1/fundandsign"
)
