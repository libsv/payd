package http

// Routes for the client api.
const (
	RouteCreatePayment   = "api/v1/payments"
	RouteAddFund         = "api/v1/funds"
	RouteGetFundsUnspent = "api/v1/funds/unspent"
	RouteTxStatus        = "api/v1/txstatus/:txid"
)
