package ipam

import "golang.zx2c4.com/wireguard/wgctrl/wgtypes"

func GenerateWGPrivateKey() (string, error) {
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return "", err
	}
	return key.String(), nil
}

func GenerateWGPublicKey(privateKey string) (string, error) {
	key, err := wgtypes.ParseKey(privateKey)
	if err != nil {
		return "", err
	}
	return key.PublicKey().String(), nil
}
