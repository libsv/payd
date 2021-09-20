package cli

import (
	"net/http"

	chttp "github.com/libsv/payd/cli/data/http"
	"github.com/libsv/payd/cli/models"
	"github.com/libsv/payd/cli/service"
	"github.com/spf13/cobra"
)

var (
	satoshis         uint64
	invoiceReference string
)

var invoiceCmd = &cobra.Command{
	Use:     "invoices",
	Aliases: []string{"invoice", "inv", "i"},
	Short:   "create or get invoices",
	Long:    "create or get invoices",
	Run:     func(cmd *cobra.Command, args []string) {},
}

var getInvoiceCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "get an invoice",
	Long:    "get an invoice",
	RunE:    get,
}

var createInvoiceCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"c"},
	Short:   "create an invoice",
	Long:    "create an invoice",
	RunE:    create,
}

var deleteInvoiceCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"d"},
	Short:   "delete an invoice",
	Long:    "delete an invoice",
	Args:    cobra.MinimumNArgs(1),
	RunE:    remove,
}

func init() {
	rootCmd.AddCommand(invoiceCmd)

	// Get invoice
	invoiceCmd.AddCommand(getInvoiceCmd)

	// Create invoice
	createInvoiceCmd.PersistentFlags().Uint64VarP(&satoshis, "satoshis", "s", 0, "invoice value")
	createInvoiceCmd.Flags().StringVarP(&invoiceReference, "reference", "r", "", "invoice reference [optional]")
	invoiceCmd.AddCommand(createInvoiceCmd)

	// Delete invoice
	invoiceCmd.AddCommand(deleteInvoiceCmd)
}

func get(cmd *cobra.Command, args []string) error {
	svc := service.NewInvoiceService(chttp.NewInvoiceAPI(&http.Client{}))
	if len(args) == 0 {
		inv, err := svc.Invoices(cmd.Context())
		if err != nil {
			return err
		}
		return printer(inv)
	}

	inv, err := svc.Invoice(cmd.Context(), models.InvoiceGetArgs{
		ID: args[0],
	})
	if err != nil {
		return err
	}

	return printer(inv)
}

func create(cmd *cobra.Command, args []string) error {
	svc := service.NewInvoiceService(chttp.NewInvoiceAPI(&http.Client{}))
	inv, err := svc.Create(cmd.Context(), models.InvoiceCreateRequest{
		Satoshis:  satoshis,
		Account:   account,
		Reference: invoiceReference,
	})
	if err != nil {
		return err
	}

	return printer(inv)
}

func remove(cmd *cobra.Command, args []string) error {
	svc := service.NewInvoiceService(chttp.NewInvoiceAPI(&http.Client{}))
	if err := svc.Delete(cmd.Context(), models.InvoiceDeleteArgs{
		ID: args[0],
	}); err != nil {
		return err
	}

	return printer(args[0] + " successfully deleted")
}
