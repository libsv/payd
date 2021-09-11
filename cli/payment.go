package cli

import (
	"net/http"

	chttp "github.com/libsv/payd/cli/data/http"
	"github.com/libsv/payd/cli/models"
	"github.com/libsv/payd/cli/service"
	"github.com/spf13/cobra"
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
	RunE:    request,
}

func init() {
	rootCmd.AddCommand(paymentCmd)

	paymentCmd.AddCommand(paymentRequestCmd)
}

func request(cmd *cobra.Command, args []string) error {
	svc := service.NewPaymentService(chttp.NewPaymentAPI(&http.Client{}))
	req, err := svc.Request(cmd.Context(), models.PaymentRequestArgs{
		ID: args[0],
	})
	if err != nil {
		return err
	}

	return printer.Print(req)
}
