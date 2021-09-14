package cli

import (
	"net/http"

	chttp "github.com/libsv/payd/cli/data/http"
	"github.com/libsv/payd/cli/models"
	"github.com/libsv/payd/cli/service"
	"github.com/spf13/cobra"
)

var (
	fundAccount string
	fundAmount  uint64
)

var fundCmd = &cobra.Command{
	Use:     "funds",
	Aliases: []string{"fund", "fun", "f"},
	Short:   "manage funds",
	Long:    "manage funds",
}

var addFundCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"a"},
	Short:   "add a fund",
	Long:    "add a fund",
	Args:    cobra.MinimumNArgs(1),
	RunE:    fundAdd,
}

var getFundCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "get funds",
	Long:    "get funds",
	RunE:    fundGet,
}

func init() {
	rootCmd.AddCommand(fundCmd)

	addFundCmd.PersistentFlags().StringVarP(&fundAccount, "account", "a", "client", "account to fund")
	fundCmd.AddCommand(addFundCmd)

	getFundCmd.PersistentFlags().StringVarP(&fundAccount, "account", "a", "client", "account to fund")
	getFundCmd.PersistentFlags().Uint64VarP(&fundAmount, "total", "t", 0, "desired amount")
	fundCmd.AddCommand(getFundCmd)
}

func fundAdd(cmd *cobra.Command, args []string) error {
	svc := service.NewFundService(chttp.NewFundAPI(&http.Client{}))
	txos, err := svc.Add(cmd.Context(), models.FundAddArgs{
		TxHex:   args[0],
		Account: fundAccount,
	})
	if err != nil {
		return err
	}

	return printer(txos)
}

func fundGet(cmd *cobra.Command, args []string) error {
	svc := service.NewFundService(chttp.NewFundAPI(&http.Client{}))
	funds, err := svc.Get(cmd.Context(), models.FundGetArgs{
		Account: fundAccount,
	})
	if err != nil {
		return err
	}

	return printer(funds)
}
