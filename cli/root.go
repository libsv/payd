package cli

import (
	"github.com/libsv/payd/cli/prnt"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	output  string
)

var rootCmd = &cobra.Command{
	Use:   "payctl",
	Short: "Interface with payd",
	Long:  "Interface with payd",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		printer = prnt.NewPrinter(prnt.Format(output))
	},
}

var printer prnt.Printer

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "version", "v", false, "verbose printing")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "output format")
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
