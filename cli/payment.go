package cli

import (
	"errors"
	"net/http"
	"net/url"

	chttp "github.com/libsv/payd/cli/data/http"
	"github.com/libsv/payd/cli/models"
	"github.com/libsv/payd/cli/service"
	"github.com/spf13/cobra"
)

var (
	requestJSON  string
	payToURL     string
	payToContext string
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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if payToURL == "" && payToContext == "" {
			return errors.New("either --pay-to-url or --pay-to-context must be provided")
		}

		if payToContext != "" {
			if ok := cfg.HasContext(payToContext); !ok {
				return ErrContextNotFound
			}

			payToURL = cfg.Contexts[payToContext].P4.URLFor()
		}

		_, err := url.Parse(payToURL)
		return err
	},
	Args: cobra.MinimumNArgs(1),
	RunE: paymentRequest,
}

func init() {
	rootCmd.AddCommand(paymentCmd)

	paymentRequestCmd.PersistentFlags().StringVarP(&payToURL, "pay-to-url", "u", "", "the payto url")
	paymentRequestCmd.PersistentFlags().StringVarP(&payToContext, "pay-to-context", "c", "", "the payto context")
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
