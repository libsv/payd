package cli

import (
	"io/ioutil"
	"os"

	"github.com/libsv/payd/cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	accountName string
	walletHost  string
	walletPort  string
)

var configCmd = &cobra.Command{
	Use:           "config",
	SilenceErrors: true,
	SilenceUsage:  true,
	Short:         "View/edit the config",
	Long:          "View/edit the config",
}

var getConfigCmd = &cobra.Command{
	Use:           "get",
	SilenceErrors: true,
	SilenceUsage:  true,
	Short:         "view the config",
	Long:          "view the config",
	RunE:          getConfig,
}

var setConfigCmd = &cobra.Command{
	Use:           "set",
	SilenceErrors: true,
	SilenceUsage:  true,
	Short:         "set config value",
	Long:          "set config value",
	RunE:          setConfig,
}

var useContextCmd = &cobra.Command{
	Use:           "usecontext",
	SilenceErrors: true,
	SilenceUsage:  true,
	Short:         "change default context",
	Long:          "change default context",
	Args:          cobra.MinimumNArgs(1),
	RunE:          useDefaultContext,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(getConfigCmd)
	configCmd.AddCommand(setConfigCmd)
	configCmd.AddCommand(useContextCmd)

	setConfigCmd.Flags().StringVarP(&accountName, "account-name", "", "", "set the account name")
	setConfigCmd.Flags().StringVarP(&walletHost, "wallet-host", "", "", "set the wallet host name")
	setConfigCmd.Flags().StringVarP(&walletPort, "wallet-port", "", "", "set the wallet port")
}

func getConfig(cmd *cobra.Command, args []string) error {
	f, err := os.Open(viper.ConfigFileUsed())
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	return printer(string(b))
}

func setConfig(cmd *cobra.Command, args []string) error {
	if accountName != "" {
		viper.Set(config.CfgAccountName.KeyFor(cfg.CurrentContext), accountName)
	}
	if walletHost != "" {
		viper.Set(config.CfgWalletHost.KeyFor(cfg.CurrentContext), walletHost)
	}
	if walletPort != "" {
		viper.Set(config.CfgWalletPort.KeyFor(cfg.CurrentContext), walletPort)
	}

	return viper.WriteConfig()
}

func useDefaultContext(cmd *cobra.Command, args []string) error {
	if ok := cfg.LoadContext(args[0]); !ok {
		return ErrContextNotFound
	}

	viper.Set(config.CfgCurrentContext, args[0])
	return viper.WriteConfig()
}
