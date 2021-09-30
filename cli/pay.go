package cli

import (
	"fmt"
	"net/http"

	chttp "github.com/libsv/payd/cli/data/http"
	"github.com/libsv/payd/cli/models"
	"github.com/libsv/payd/cli/service"
	"github.com/spf13/cobra"
)

var payCmd = &cobra.Command{
	Use:     "pay",
	Aliases: []string{"pa", "p"},
	Short:   "pay",
	Long:    "send satoshis to address",
	Args:    cobra.MinimumNArgs(1),
	PreRunE: paymentValidator,
	RunE:    pay,
}

func init() {
	rootCmd.AddCommand(payCmd)
	payCmd.PersistentFlags().StringVarP(&payToURL, "pay-to-url", "u", "", "the payto url")
	payCmd.PersistentFlags().StringVarP(&payToContext, "pay-to-context", "c", "", "the payto context")
}

func pay(cmd *cobra.Command, args []string) error {
	svc := service.NewPayService(chttp.NewPayAPI(&http.Client{}, cfg.Payd))
	ack, err := svc.Pay(cmd.Context(), models.PayRequest{
		PayToURL: fmt.Sprintf("%s/api/v1/payment/%s", payToURL, args[0]),
	})
	if err != nil {
		return err
	}

	return printer(ack)
}
