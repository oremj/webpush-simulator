package pushclient

import (
	"crypto/elliptic"
	"crypto/rand"
)

var curve = elliptic.P256()

type ClientEncryption struct {
	pub  []byte
	priv []byte
	auth []byte
}

func genAuth() []byte {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
	return buf
}

func genKeyPair() ([]byte, []byte) {
	priv, x, y, err := elliptic.GenerateKey(curve, rand.Reader)
	if err != nil {
		panic(err)
	}
	pub := elliptic.Marshal(curve, x, y)

	return pub, priv
}

func NewClientEncryption() *ClientEncryption {
	pub, priv := genKeyPair()
	return &ClientEncryption{
		auth: genAuth(),
		pub:  pub,
		priv: priv,
	}
}

func (c *ClientEncryption) PubKey() []byte {
	return c.pub
}

func (c *ClientEncryption) AuthKey() []byte {
	return c.auth
}
