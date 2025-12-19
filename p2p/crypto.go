package p2p

import (
	"crypto/rand"
	"math/big"
)

type CardKeys struct {
	EncryptionKey 	*big.Int
	DecryptionKey 	*big.Int 
	Prime 			*big.Int
}

func GenerateCardKeys(sharedPrime *big.Int) (*CardKeys, error) {
	phi := new(big.Int).Sub(sharedPrime, big.NewInt(1))
	for {
		e, err := rand.Int(rand.Reader, phi)
		if err != nil{
			return nil, err
		}
		if e.Cmp(big.NewInt(1)) <= 0 {
			continue
		}
		gcd := new(big.Int).GCD(nil, nil, e, phi)
		if gcd.Cmp(big.NewInt(1)) == 0 {
			d := new(big.Int).ModInverse(e, phi)
			if d != nil {
				return &CardKeys{
					EncryptionKey: e,
					DecryptionKey: d,
					Prime: sharedPrime,
				}, nil
			}
		}
	}
}

func (k *CardKeys) Encrypt(data []byte) []byte {
	m := new(big.Int).SetBytes(data)
	c := new(big.Int).Exp(m, k.EncryptionKey, k.Prime)
	return c.Bytes()
}

func (k *CardKeys) Decrypt(data []byte) []byte {
	c := new(big.Int).SetBytes(data)
	m := new(big.Int).Exp(c, k.DecryptionKey, k.Prime)
	return m.Bytes()
}