package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ConfigValidateDB(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		cfg *Config
		err error
	}{
		"valid db config (sqlite) should return no errors": {
			cfg: &Config{
				Db: &Db{
					Type: "sqlite",
				},
			},
			err: nil,
		}, "valid db config (postgres) should return no errors": {
			cfg: &Config{
				Db: &Db{
					Type: "postgres",
				},
			},
			err: nil,
		}, "valid db config (mysql) should return no errors": {
			cfg: &Config{
				Db: &Db{
					Type: "mysql",
				},
			},
			err: nil,
		}, "invalid db config should return no errors": {
			cfg: &Config{
				Db: &Db{
					Type: "mydb",
				},
			},
			err: errors.New("[db.type: value mydb failed to meet requirements]"),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := test.cfg.Validate()
			if test.err == nil {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, test.err.Error())

		})
	}
}

func Test_ConfigValidateNetwork(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		cfg *Config
		err error
	}{
		"valid deployment network config (mainnet) should return no errors": {
			cfg: &Config{
				Wallet: &Wallet{
					Network: NetworkMainet,
				},
			},
			err: nil,
		}, "valid deployment network config (testnet) should return no errors": {
			cfg: &Config{
				Wallet: &Wallet{
					Network: NetworkTestnet,
				},
			},
			err: nil,
		}, "valid deployment network config (stn) should return no errors": {
			cfg: &Config{
				Wallet: &Wallet{
					Network: NetworkSTN,
				},
			},
			err: nil,
		}, "valid deployment network config (regtest) should return no errors": {
			cfg: &Config{
				Wallet: &Wallet{
					Network: NetworkRegtest,
				},
			},
			err: nil,
		}, "deployment network type within other word should fail": {
			cfg: &Config{
				Wallet: &Wallet{
					Network: "btestneth",
				},
			},
			err: errors.New("[wallet.network: value btestneth failed to meet requirements]"),
		}, "invalid deployment network config should error": {
			cfg: &Config{
				Wallet: &Wallet{
					Network: "blah",
				},
			},
			err: errors.New("[wallet.network: value blah failed to meet requirements]"),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := test.cfg.Validate()
			if test.err == nil {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, test.err.Error())

		})
	}
}
