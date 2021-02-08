package wallet

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bitcoinsv/bsvd/bsvec"
	"github.com/bitcoinsv/bsvd/chaincfg"
	"github.com/bitcoinsv/bsvutil/hdkeychain"
	"github.com/pkg/errors"
)

var (
	numericPlusTick = regexp.MustCompile(`^[0-9]+'{0,1}$`)
)

// CreatePrivateKey creates a extended private key for a keyname
func CreatePrivateKey(keyname string) error { // get keyname from settings in caller

	// TODO: open conn to sqlite db and defer

	// TODO: check if key exists

	seed, err := hdkeychain.GenerateSeed(hdkeychain.RecommendedSeedLen)
	if err != nil {
		return errors.Wrap(err, "failed to generate seed")
	}

	mainNet := false // TODO: get get config

	var chain *chaincfg.Params

	if mainNet {
		chain = &chaincfg.MainNetParams
	} else {
		chain = &chaincfg.TestNet3Params
	}

	xprv, err := hdkeychain.NewMaster(seed, chain)
	if err != nil {
		return errors.Wrap(err, "failed to create master node for given seed and chain")
	}

	// TODO: insert xprv into db
	fmt.Println(xprv)

	return nil
}

// GetPrivateKey returns the extended private key for a keyname
func GetPrivateKey(keyname string) (*hdkeychain.ExtendedKey, error) {

	xpriv := "" // TODO: get from db

	key, err := hdkeychain.NewKeyFromString(xpriv)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// DeriveChildFromKey derives a child extended private key from an extended private key
func DeriveChildFromKey(startingKey hdkeychain.ExtendedKey, derivationPath string) (*hdkeychain.ExtendedKey, error) { // TODO: check startingKey not pointer

	// This method does not appear to be thread safe so we pass the starting key by value and then point to that...
	key := &startingKey

	if derivationPath != "" {
		children := strings.Split(derivationPath, "/")

		for _, child := range children {
			if !isValidSegment(child) {
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

func isValidSegment(child string) bool {
	return numericPlusTick.MatchString(child)
}

func getChildInt(child string) (uint32, error) {
	var suffix uint32

	if strings.HasSuffix(child, "'") {
		child = strings.TrimRight(child, "'")
		suffix = 2147483648 // 2^32
	}

	t, err := strconv.Atoi(child)
	if err != nil {
		return 0, errors.Wrap(err, "getChildInt: "+child)
	}

	return uint32(t) + suffix, nil
}

// PrivFromXPrv returns an ECDSA private key from an extended private key
func PrivFromXPrv(xprv *hdkeychain.ExtendedKey) (*bsvec.PrivateKey, error) {
	return xprv.ECPrivKey()
}

// PubFromXPrv returns an ECDSA public key from an extended private key
func PubFromXPrv(xprv *hdkeychain.ExtendedKey) ([]byte, error) {
	pub, err := xprv.ECPubKey()
	if err != nil {
		return nil, err
	}

	return pub.SerializeCompressed(), nil
}
