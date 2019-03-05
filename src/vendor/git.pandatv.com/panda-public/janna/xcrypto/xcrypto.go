package xcrypto

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"errors"
)

var (
	ErrNotFullBlocks = errors.New("not full blocks")
	ErrDecodeFail    = errors.New("decode fail")
)

type Crypter interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
}

type DesCrypter struct {
	cb  cipher.Block
	key []byte
	bs  int
}

func NewDes(key []byte) (*DesCrypter, error) {
	b, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	dc := new(DesCrypter)
	dc.bs = des.BlockSize
	dc.cb = b
	dc.key = key

	return dc, nil
}

func (this *DesCrypter) Encrypt(src []byte) ([]byte, error) {
	bsrc := this.padding(src)
	bdst := make([]byte, len(bsrc))
	cipher.NewCBCEncrypter(this.cb, this.key).CryptBlocks(bdst, bsrc)

	return bdst, nil
}

func (this *DesCrypter) Decrypt(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return []byte{}, nil
	}
	if len(src)%this.bs != 0 {
		return nil, ErrNotFullBlocks
	}

	dst := make([]byte, len(src))
	cipher.NewCBCDecrypter(this.cb, this.key).CryptBlocks(dst, src)

	bdst, err := this.unpadding(dst)
	if err != nil {
		return nil, err
	}

	return bdst, nil
}

func (this *DesCrypter) padding(src []byte) []byte {
	l := len(src)
	n := l % this.bs
	src = append(src, bytes.Repeat([]byte{byte(this.bs - n)}, this.bs-n)...)

	return src
}

func (this *DesCrypter) unpadding(src []byte) ([]byte, error) {
	length := len(src)
	n := int(src[length-1])
	if n > this.bs || (length < n) {
		return nil, ErrDecodeFail
	}

	src = src[:length-n]
	return src, nil
}
