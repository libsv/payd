package http

const (
	RoutePaymentRequest = "r/:paymentID"
	RoutePayment        = "payment/:paymentID"

	RouteInvoice  = "v1/invoices/:paymentID"
	RouteInvoices = "v1/invoices"
	RouteBalance  = "v1/balance"
)
