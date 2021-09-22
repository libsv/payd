package cli

import (
	"io/ioutil"
	"os"

	"github.com/libsv/payd/cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	defaultAccount string
	walletHost     string
	walletPort     string
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

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(getConfigCmd)
	configCmd.AddCommand(setConfigCmd)

	setConfigCmd.Flags().StringVarP(&defaultAccount, "default-account", "", "", "set the default account")
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
	if defaultAccount != "" {
		viper.Set(config.CfgAccountName.Key(), defaultAccount)
	}

	if walletHost != "" {
		viper.Set(config.CfgWalletHost.Key(), walletHost)
	}
	if walletPort != "" {
		viper.Set(config.CfgWalletPort.Key(), walletPort)
	}

	return viper.WriteConfig()
}
