package internal

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/libsv/go-bc/spv"
	"github.com/pkg/errors"
	"github.com/theflyingcodr/sockets/client"
	"github.com/tonicpow/go-minercraft"

	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
	dataHttp "github.com/libsv/payd/data/http"
	"github.com/libsv/payd/data/mapi"
	dsoc "github.com/libsv/payd/data/sockets"
	paydSQL "github.com/libsv/payd/data/sqlite"
	"github.com/libsv/payd/service"
)

// RestDeps contains all dependencies used for the rest client.
type RestDeps struct {
	DestinationService payd.DestinationsService
	PaymentService     payd.PaymentsService
	PayService         payd.PayService
	EnvelopeService    payd.EnvelopeService
	InvoiceService     payd.InvoiceService
	BalanceService     payd.BalanceService
	ProofService       payd.ProofsService
	OwnerService       payd.OwnerService
}

// SetupRestDeps will setup dependencies used in the rest server.
func SetupRestDeps(cfg *config.Config, db *sqlx.DB) *RestDeps {
	mapiCli, err := minercraft.NewClient(nil, nil, []*minercraft.Miner{
		{
			Name:  cfg.Mapi.MinerName,
			Token: cfg.Mapi.Token,
			URL:   cfg.Mapi.URL,
		},
	})
	if err != nil {
		log.Fatal(mapiCli)
	}
	sqlLiteStore := paydSQL.NewSQLiteStore(db)
	mapiStore := mapi.NewMapi(cfg.Mapi, mapiCli)
	spvv, err := spv.NewPaymentVerifier(dataHttp.NewHeaderSVConnection(&http.Client{Timeout: time.Duration(cfg.HeadersClient.Timeout) * time.Second}, cfg.HeadersClient.Address))
	if err != nil {
		log.Fatalf("failed to create spv client %s", err)
	}

	spvc, err := spv.NewEnvelopeCreator(sqlLiteStore, sqlLiteStore)
	if err != nil {
		log.Fatalf("failed to create spv verifier %s", err)
	}

	seedSvc := service.NewSeedService()
	privKeySvc := service.NewPrivateKeys(sqlLiteStore, cfg.Wallet.Network == "mainnet")
	destSvc := service.NewDestinationsService(cfg.Wallet, privKeySvc, sqlLiteStore, sqlLiteStore, sqlLiteStore, mapiStore, seedSvc)
	paymentSvc := service.NewPayments(spvv, sqlLiteStore, sqlLiteStore, sqlLiteStore, &paydSQL.Transacter{}, mapiStore, mapiStore, sqlLiteStore)
	envSvc := service.NewEnvelopes(privKeySvc, sqlLiteStore, sqlLiteStore, sqlLiteStore, seedSvc, spvc)
	paySvc := service.NewPayService(dataHttp.NewP4(&http.Client{Timeout: time.Duration(cfg.P4.Timeout) * time.Second}), envSvc, cfg.Server)
	invoiceSvc := service.NewInvoice(cfg.Server, cfg.Wallet, sqlLiteStore, destSvc, &paydSQL.Transacter{}, service.NewTimestampService())
	balanceSvc := service.NewBalance(sqlLiteStore)
	proofSvc := service.NewProofsService(sqlLiteStore)
	ownerSvc := service.NewOwnerService(sqlLiteStore)

	// create master private key if it doesn't exist
	if err = privKeySvc.Create(context.Background(), "masterkey"); err != nil {
		log.Fatal(errors.Wrap(err, "failed to create master key"))
	}

	return &RestDeps{
		DestinationService: destSvc,
		PaymentService:     paymentSvc,
		PayService:         paySvc,
		EnvelopeService:    envSvc,
		InvoiceService:     invoiceSvc,
		BalanceService:     balanceSvc,
		ProofService:       proofSvc,
		OwnerService:       ownerSvc,
	}
}

// SocketDeps contains all dependencies of the socket server.
type SocketDeps struct {
	DestinationService    payd.DestinationsService
	PaymentService        payd.PaymentsService
	PayService            payd.PayService
	EnvelopeService       payd.EnvelopeService
	InvoiceService        payd.InvoiceService
	BalanceService        payd.BalanceService
	ProofService          payd.ProofsService
	OwnerService          payd.OwnerService
	PaymentRequestService payd.PaymentRequestService
	ConnectService        payd.ConnectService
	TransactionService    payd.TransactionService
}

// SetupSocketDeps will setup dependencies used in the socket server.
func SetupSocketDeps(cfg *config.Config, db *sqlx.DB, c *client.Client) *SocketDeps {
	mapiCli, err := minercraft.NewClient(nil, nil, []*minercraft.Miner{
		{
			Name:  cfg.Mapi.MinerName,
			Token: cfg.Mapi.Token,
			URL:   cfg.Mapi.URL,
		},
	})
	if err != nil {
		log.Fatal(mapiCli)
	}
	sqlLiteStore := paydSQL.NewSQLiteStore(db)
	mapiStore := mapi.NewMapi(cfg.Mapi, mapiCli)
	spvv, err := spv.NewPaymentVerifier(dataHttp.NewHeaderSVConnection(&http.Client{Timeout: time.Duration(cfg.HeadersClient.Timeout) * time.Second}, cfg.HeadersClient.Address))
	if err != nil {
		log.Fatalf("failed to create spv client %s", err)
	}

	spvc, err := spv.NewEnvelopeCreator(sqlLiteStore, sqlLiteStore)
	if err != nil {
		log.Fatalf("failed to create spv verifier %s", err)
	}

	seedSvc := service.NewSeedService()
	privKeySvc := service.NewPrivateKeys(sqlLiteStore, cfg.Wallet.Network == "mainnet")
	destSvc := service.NewDestinationsService(cfg.Wallet, privKeySvc, sqlLiteStore, sqlLiteStore, sqlLiteStore, mapiStore, seedSvc)
	paymentSvc := service.NewPayments(spvv, sqlLiteStore, sqlLiteStore, sqlLiteStore, &paydSQL.Transacter{}, mapiStore, mapiStore, sqlLiteStore)
	envSvc := service.NewEnvelopes(privKeySvc, sqlLiteStore, sqlLiteStore, sqlLiteStore, seedSvc, spvc)
	paySvc := service.NewPayChannel(dsoc.NewPaymentChannel(*cfg.Socket, c))
	invoiceSvc := service.NewInvoice(cfg.Server, cfg.Wallet, sqlLiteStore, destSvc, &paydSQL.Transacter{}, service.NewTimestampService())
	balanceSvc := service.NewBalance(sqlLiteStore)
	proofSvc := service.NewProofsService(sqlLiteStore)
	ownerSvc := service.NewOwnerService(sqlLiteStore)
	paymentReqSvc := service.NewPaymentRequest(cfg.Wallet, destSvc, sqlLiteStore)
	connectService := service.NewConnect(dsoc.NewConnect(cfg.P4, c), invoiceSvc)
	transactionService := service.NewTransactions(&paydSQL.Transacter{}, sqlLiteStore, sqlLiteStore, sqlLiteStore)

	// create master private key if it doesn't exist
	if err = privKeySvc.Create(context.Background(), "masterkey"); err != nil {
		log.Fatal(errors.Wrap(err, "failed to create master key"))
	}

	return &SocketDeps{
		DestinationService:    destSvc,
		PaymentService:        paymentSvc,
		PayService:            paySvc,
		EnvelopeService:       envSvc,
		InvoiceService:        invoiceSvc,
		BalanceService:        balanceSvc,
		ProofService:          proofSvc,
		OwnerService:          ownerSvc,
		PaymentRequestService: paymentReqSvc,
		ConnectService:        connectService,
		TransactionService:    transactionService,
	}
}
