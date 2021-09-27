package cli

import (
	"net/http"

	chttp "github.com/libsv/payd/cli/data/http"
	"github.com/libsv/payd/cli/service"
	"github.com/spf13/cobra"
)

var balanceCmd = &cobra.Command{
	Use:     "balance",
	Aliases: []string{"balanc", "balan", "bala", "bal", "ba", "b"},
	Short:   "view wallet balance",
	Long:    "view wallet balance",
	RunE:    getBalance,
}

func init() {
	rootCmd.AddCommand(balanceCmd)
}

func getBalance(cmd *cobra.Command, args []string) error {
	svc := service.NewBalanceService(chttp.NewBalanceAPI(&http.Client{}, cfg.Payd))
	bal, err := svc.Balance(cmd.Context())
	if err != nil {
		return err
	}

	return printer(bal)
}
