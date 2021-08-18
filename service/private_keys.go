package service

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bk/chaincfg"
	gopayd "github.com/libsv/payd"
	"github.com/pkg/errors"
)

type privateKey struct {
	store           gopayd.PrivateKeyReaderWriter
	useMainNet      bool
	numericPlusTick *regexp.Regexp
}

// NewPrivateKeys will setup and return a new PrivateKey service.
func NewPrivateKeys(store gopayd.PrivateKeyReaderWriter, useMainNet bool) *privateKey {
	return &privateKey{
		store:           store,
		useMainNet:      useMainNet,
		numericPlusTick: regexp.MustCompile(`^[0-9]+'{0,1}$`),
	}
}

// Create creates a extended private key for a keyName.
func (svc *privateKey) Create(ctx context.Context, keyName string) error { // get keyname from settings in caller
	key, err := svc.store.PrivateKey(ctx, gopayd.KeyArgs{Name: keyName})
	if err != nil {
		return errors.Wrapf(err, "failed to get key %s by name", keyName)
	}
	if key != nil {
		return nil
	}
	seed, err := bip32.GenerateSeed(bip32.RecommendedSeedLen)
	if err != nil {
		return errors.Wrap(err, "failed to generate seed")
	}
	chain := &chaincfg.TestNet
	if svc.useMainNet {
		chain = &chaincfg.MainNet
	}
	xprv, err := bip32.NewMaster(seed, chain)
	if err != nil {
		return errors.Wrap(err, "failed to create master node for given seed and chain")
	}
	if _, err := svc.store.CreatePrivateKey(ctx, gopayd.PrivateKey{
		Name: keyName,
		Xprv: xprv.String(),
	}); err != nil {
		return errors.Wrap(err, "failed to create private key")
	}
	return nil
}

// PrivateKey returns the extended private key for a keyname.
func (svc *privateKey) PrivateKey(ctx context.Context, keyName string) (*bip32.ExtendedKey, error) {
	key, err := svc.store.PrivateKey(ctx, gopayd.KeyArgs{Name: keyName})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get key %s by name", keyName)
	}
	if key == nil {
		return nil, errors.Wrap(err, "key not found")
	}

	xKey, err := bip32.NewKeyFromString(key.Xprv)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get extended key from xpriv")
	}
	return xKey, nil
}

// DeriveChildFromKey will create a private key derived from a parent extended private key at the given derivationPath.
func (svc *privateKey) DeriveChildFromKey(startingKey *bip32.ExtendedKey, derivationPath string) (*bip32.ExtendedKey, error) { // TODO: check startingKey not pointer
	key := startingKey
	if derivationPath != "" {
		children := strings.Split(derivationPath, "/")
		for _, child := range children {
			if !svc.numericPlusTick.MatchString(child) {
				return nil, errors.Wrap(errors.New("deriveChildFromKey failed"), fmt.Sprintf("invalid path: %q", derivationPath))
			}
			childInt, err := getChildInt(child)
			if err != nil {
				return nil, errors.Wrap(err, "deriveChildFromKey failed")
			}
			var childErr error
			key, childErr = key.Child(childInt)
			if childErr != nil {
				return nil, errors.Wrap(childErr, fmt.Sprintf("deriveChildFromKey: key.Child: %v", childInt))
			}
		}
	}
	return key, nil
}

func getChildInt(child string) (uint32, error) {
	var suffix uint32
	if strings.HasSuffix(child, "'") {
		child = strings.TrimRight(child, "'")
		suffix = 2147483648 // 2^32
	}
	t, err := strconv.ParseUint(child, 10, 32)
	if err != nil {
		return 0, errors.Wrap(err, "getChildInt: "+child)
	}
	return uint32(t) + suffix, nil
}

// PubFromXPrv returns an ECDSA public key from an extended private key.
func (svc *privateKey) PubFromXPrv(xprv *bip32.ExtendedKey) (*bec.PublicKey, error) {
	pub, err := xprv.ECPubKey()
	if err != nil {
		return nil, err
	}
	return pub, nil
}
