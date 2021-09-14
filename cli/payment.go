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

var (
	requestJSON string
)

var paymentCmd = &cobra.Command{
	Use:     "payment",
	Aliases: []string{"p"},
	Short:   "payments",
	Long:    "payments",
}

var paymentRequestCmd = &cobra.Command{
	Use:     "request",
	Aliases: []string{"req"},
	Short:   "request a payment",
	Long:    "request a payment",
	Args:    cobra.MinimumNArgs(1),
	RunE:    paymentRequest,
}

var paymentSendCmd = &cobra.Command{
	Use:     "send",
	Aliases: []string{"sen", "s"},
	Short:   "send a payment",
	Long:    "send a payment",
	RunE:    paymentSend,
}

func init() {
	rootCmd.AddCommand(paymentCmd)

	paymentCmd.AddCommand(paymentRequestCmd)

	paymentSendCmd.PersistentFlags().StringVarP(&requestJSON, "request", "r", "", "the payment to send")
	paymentCmd.AddCommand(paymentSendCmd)
}

func paymentRequest(cmd *cobra.Command, args []string) error {
	svc := service.NewPaymentService(chttp.NewPaymentAPI(&http.Client{}), nil, nil, nil)
	req, err := svc.Request(cmd.Context(), models.PaymentRequestArgs{
		ID: args[0],
	})
	if err != nil {
		return err
	}

	return printer(req)
}

func paymentSend(cmd *cobra.Command, args []string) error {
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

	svc := service.NewPaymentService(
		chttp.NewPaymentAPI(&http.Client{}),
		service.NewFundService(chttp.NewFundAPI(&http.Client{})),
		chttp.NewSignerAPI(&http.Client{}),
		spvb,
	)
	resp, err := svc.Send(cmd.Context(), models.PaymentSendArgs{
		PaymentRequest: payReq,
	})
	if err != nil {
		return err
	}

	return printer(resp)
}
