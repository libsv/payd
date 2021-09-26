package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/libsv/payd/cli/config"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	setAsCurrentContext bool
	accountName         string
	paydHost            string
	paydPort            string
	p4Host              string
	p4Port              string
)

var configCmd = &cobra.Command{
	Use:           "config",
	SilenceErrors: true,
	SilenceUsage:  true,
	Short:         "View/edit the config",
	Long:          "View/edit the config",
}

var printConfigCmd = &cobra.Command{
	Use:           "print",
	SilenceErrors: true,
	SilenceUsage:  true,
	Short:         "print the config",
	Long:          "print the config",
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

var addContextCmd = &cobra.Command{
	Use:           "addcontext",
	SilenceErrors: true,
	SilenceUsage:  true,
	Short:         "add a context",
	Long:          "add a context",
	Args:          cobra.MinimumNArgs(1),
	RunE:          addContext,
}

var useContextCmd = &cobra.Command{
	Use:           "usecontext",
	SilenceErrors: true,
	SilenceUsage:  true,
	Short:         "change default context",
	Long:          "change default context",
	Args:          cobra.MinimumNArgs(1),
	PreRunE:       preRunContextExists,
	RunE:          useDefaultContext,
}

var deleteContextCmd = &cobra.Command{
	Use:           "deletecontext",
	SilenceErrors: true,
	SilenceUsage:  true,
	Short:         "delete a context",
	Long:          "delete a context",
	Args:          cobra.MinimumNArgs(1),
	PreRunE:       preRunContextExists,
	RunE:          deleteContext,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(printConfigCmd)
	configCmd.AddCommand(setConfigCmd)
	configCmd.AddCommand(useContextCmd)
	configCmd.AddCommand(addContextCmd)
	configCmd.AddCommand(deleteContextCmd)

	addContextCmd.Flags().BoolVarP(&setAsCurrentContext, "apply", "", false, "set the account name")
	addContextCmd.Flags().StringVarP(&accountName, "account-name", "", "", "set the account name")
	addContextCmd.Flags().StringVarP(&paydHost, "payd-host", "", "", "set the payd host name")
	addContextCmd.Flags().StringVarP(&paydPort, "payd-port", "", "", "set the payd port")
	addContextCmd.Flags().StringVarP(&p4Host, "p4-host", "", "", "set the p4 host name")
	addContextCmd.Flags().StringVarP(&p4Port, "p4-port", "", "", "set the p4 port")

	setConfigCmd.Flags().StringVarP(&accountName, "account-name", "", "", "set the account name")
	setConfigCmd.Flags().StringVarP(&paydHost, "payd-host", "", "", "set the payd host name")
	setConfigCmd.Flags().StringVarP(&paydPort, "payd-port", "", "", "set the payd port")
	setConfigCmd.Flags().StringVarP(&p4Host, "p4-host", "", "", "set the p4 host name")
	setConfigCmd.Flags().StringVarP(&p4Port, "p4-port", "", "", "set the p4 port")
}

func preRunContextExists(cmd *cobra.Command, args []string) error {
	if ok := cfg.HasContext(args[0]); !ok {
		return ErrContextNotFound
	}

	return nil
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
	setIfNotZero(config.CfgAccountName.KeyFor(cfg.CurrentContext), accountName)
	setIfNotZero(config.CfgPaydHost.KeyFor(cfg.CurrentContext), paydHost)
	setIfNotZero(config.CfgPaydPort.KeyFor(cfg.CurrentContext), paydPort)
	setIfNotZero(config.CfgP4Host.KeyFor(cfg.CurrentContext), p4Host)
	setIfNotZero(config.CfgP4Port.KeyFor(cfg.CurrentContext), p4Port)

	return viper.WriteConfig()
}

func addContext(cmd *cobra.Command, args []string) error {
	name := args[0]
	if ok := cfg.HasContext(name); ok {
		return ErrContextAlreadyExists
	}

	viper.Set(config.CfgAccountName.KeyFor(name), defaultIfZero(accountName, name))

	viper.Set(config.CfgPaydHost.KeyFor(name), defaultIfZero(paydHost, "payd:8443"))
	viper.Set(config.CfgPaydPort.KeyFor(name), defaultIfZero(paydPort, ":8443"))

	viper.Set(config.CfgP4Host.KeyFor(name), defaultIfZero(p4Host, "p4:8445"))
	viper.Set(config.CfgP4Port.KeyFor(name), defaultIfZero(p4Port, ":8445"))

	if setAsCurrentContext {
		viper.Set(config.CfgCurrentContext, name)
	}

	return viper.WriteConfig()
}

func deleteContext(cmd *cobra.Command, args []string) error {
	keys := viper.AllKeys()
	home, err := homedir.Dir()
	cobra.CheckErr(err)

	v := viper.New()
	v.AddConfigPath(home)
	v.SetConfigName(".payctl")
	v.SetConfigType("yml")
	if len(cfg.Contexts) == 1 {
		return v.WriteConfig()
	}

	deleteCurrentContext := viper.GetString(config.CfgCurrentContext) == args[0]
	if deleteCurrentContext {
		for k := range cfg.Contexts {
			if k != args[0] {
				v.Set(config.CfgCurrentContext, k)
				fmt.Println("current.context change to", k)
				break
			}
		}
	}

	prefix := fmt.Sprintf("contexts.%s", args[0])
	for _, k := range keys {
		if strings.HasPrefix(k, prefix) {
			continue
		}

		if deleteCurrentContext && k == config.CfgCurrentContext {
			continue
		}

		v.Set(k, viper.GetString(k))
	}

	return v.WriteConfig()
}

func useDefaultContext(cmd *cobra.Command, args []string) error {
	if ok := cfg.LoadContext(args[0]); !ok {
		return ErrContextNotFound
	}

	viper.Set(config.CfgCurrentContext, args[0])
	return viper.WriteConfig()
}

func defaultIfZero(value string, def string) string {
	if value == "" {
		return def
	}
	return value
}

func setIfNotZero(key string, value string) {
	if value != "" {
		viper.Set(key, value)
	}
}
