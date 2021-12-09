package sockets

// Routes contain the unique keys for socket messages used in the payment protocol.
const (
	RoutePayment                = "payment"
	RoutePaymentACK             = "payment.ack"
	RouteProofCreate            = "proof.create"
	RoutePaymentRequestCreate   = "paymentrequest.create"
	RoutePaymentRequestResponse = "paymentrequest.response"
)

// Common headers for sockets.
const (
	// HeaderOrigin identifies this client by id, used so the client doesn't handle its own messages.
	HeaderOrigin = "X-Origin-ID"
)
