package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/libsv/go-bc/spv"
	chttp "github.com/libsv/payd/cli/data/http"
	"github.com/libsv/payd/cli/data/regtest"
	"github.com/libsv/payd/cli/models"
	"github.com/libsv/payd/cli/service"
	"github.com/spf13/cobra"
)

var fundCmd = &cobra.Command{
	Use:     "fund",
	Aliases: []string{"fun", "fu", "f"},
	Short:   "fund a wallet from the node",
	Long:    "fund a wallet from the node",
	RunE:    fund,
}

var autoFundCmd = &cobra.Command{
	Use:     "autofund",
	Aliases: []string{"autofun", "autof", "aufu", "af"},
	Short:   "auto fund the current context with a specified amount of satoshis",
	Long:    "auto fund the current context with a specified amount of satoshis",
	RunE:    autoFund,
}

func init() {
	rootCmd.AddCommand(fundCmd)
	rootCmd.AddCommand(autoFundCmd)

	fundCmd.Flags().StringVarP(&payToURL, "pay-to", "p", "", "to pay to")

	autoFundCmd.Flags().Uint64VarP(&satoshis, "satoshis", "s", 0, "satoshis to fund")
}

func fund(cmd *cobra.Command, args []string) error {
	var rdr io.Reader = os.Stdin
	if requestJSON != "" {
		rdr = bytes.NewBufferString(requestJSON)
	}

	var payReq models.PaymentRequest
	if err := json.NewDecoder(rdr).Decode(&payReq); err != nil {
		return err
	}

	rt := regtest.NewRegtest(&http.Client{})
	spvb, err := spv.NewEnvelopeCreator(service.NewTxService(rt), service.NewMerkleProofStore(rt))
	if err != nil {
		return err
	}
	svc := service.NewFundService(rt, chttp.NewPaymentAPI(&http.Client{}), spvb)

	ack, err := svc.Fund(cmd.Context(), payReq)
	if err != nil {
		return err
	}

	return printer(ack)
}

func autoFund(cmd *cobra.Command, args []string) error {
	rt := regtest.NewRegtest(&http.Client{})
	invSvc := service.NewInvoiceService(chttp.NewInvoiceAPI(&http.Client{}, cfg.Payd))
	paySvc := service.NewPaymentService(chttp.NewPaymentAPI(&http.Client{}), nil)

	ctx := cmd.Context()

	addr, err := rt.GetNewAddress(ctx)
	if err != nil {
		return err
	}

	if _, err = rt.SendToAddress(ctx, *addr.Result, 1); err != nil {
		return err
	}

	if _, err = rt.Generate(ctx, 1); err != nil {
		return err
	}

	reference := "autofund"
	description := "created via autofund"

	inv, err := invSvc.Create(ctx, models.InvoiceCreateRequest{
		Satoshis:    satoshis,
		Reference:   &reference,
		Description: &description,
	})
	if err != nil {
		return err
	}

	payReq, err := paySvc.Request(ctx, models.PaymentRequestArgs{
		ID:    inv.PaymentID,
		PayTo: cfg.P4.URLFor(),
	})
	if err != nil {
		return err
	}

	bb, err := json.Marshal(payReq)
	if err != nil {
		return err
	}

	requestJSON = string(bb)

	return fund(cmd, args)
}
