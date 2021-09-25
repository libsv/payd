package cli

import (
	"net/http"

	chttp "github.com/libsv/payd/cli/data/http"
	"github.com/libsv/payd/cli/models"
	"github.com/libsv/payd/cli/service"
	"github.com/spf13/cobra"
)

var (
	requestJSON string
	payToURL    string
	tx          string
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

func init() {
	rootCmd.AddCommand(paymentCmd)

	paymentRequestCmd.PersistentFlags().StringVarP(&payToURL, "pay-to", "u", "", "the payto url")
	paymentCmd.AddCommand(paymentRequestCmd)
}

func paymentRequest(cmd *cobra.Command, args []string) error {
	svc := service.NewPaymentService(chttp.NewPaymentAPI(&http.Client{}), nil)
	req, err := svc.Request(cmd.Context(), models.PaymentRequestArgs{
		ID:    args[0],
		PayTo: payToURL,
	})
	if err != nil {
		return err
	}

	return printer(req)
}
