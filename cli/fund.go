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

func init() {
	rootCmd.AddCommand(fundCmd)
	fundCmd.Flags().StringVarP(&payToURL, "pay-to", "p", "", "to pay to")
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
