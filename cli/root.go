package cli

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"

	"github.com/libsv/payd/cli/config"
	"github.com/libsv/payd/cli/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	verbose      bool
	outputFormat string
	account      string
)

var (
	printer output.PrintFunc
	cfg     *config.Config = config.NewConfig()
)

var rootCmd = &cobra.Command{
	Use:           "payctl",
	SilenceErrors: true,
	SilenceUsage:  true,
	Short:         "Interface with payd",
	Long:          "Interface with payd",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		printer = output.NewPrinter(output.Format(outputFormat))
	},
}

func init() {
	initConfig()

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose printing")
	rootCmd.PersistentFlags().StringVarP(&account, "account", "a", "", "account name")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "output format")
}

func initConfig() {
	home, err := homedir.Dir()
	cobra.CheckErr(err)

	viper.AddConfigPath(home)
	viper.SetConfigName(".payctl")
	viper.SetConfigType("yml")

	viper.ReadInConfig()

	cfg.WithWallet().
		WithAccount()

	viper.SafeWriteConfig()
}

// Execute the command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
