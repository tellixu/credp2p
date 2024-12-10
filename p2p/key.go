package p2p

import (
	"bytes"
	"encoding/pem"
	"errors"
	"github.com/libp2p/go-libp2p/core/crypto"
)

var (
	// SignVerifyErr 签名失败
	SignVerifyErr = errors.New("sign verify err")
)

type CryptoKeyInfo struct {
	privateKey crypto.PrivKey
}

func NewCryptoKeyInfo(privateKey crypto.PrivKey) *CryptoKeyInfo {
	return &CryptoKeyInfo{privateKey: privateKey}
}

// GetPrivateKey 获取节点的私钥
func (k *CryptoKeyInfo) GetPrivateKey() crypto.PrivKey {
	return k.privateKey
}

// GetPublicKey 获取节点的公钥
func (k *CryptoKeyInfo) GetPublicKey() crypto.PubKey {
	return k.privateKey.GetPublic()
}

// Sign 私钥签名
func (k *CryptoKeyInfo) Sign(data []byte) ([]byte, error) {
	return k.privateKey.Sign(data)
}

func (k *CryptoKeyInfo) UnMarshalPrivateKey(privateKey []byte) (crypto.PrivKey, error) {
	return crypto.UnmarshalRsaPrivateKey(privateKey)
}

func (k *CryptoKeyInfo) UnMarshalPrivateKeyByPem(keyPem []byte) (crypto.PrivKey, error) {
	block, _ := pem.Decode(keyPem)
	if block == nil {

		return nil, errors.New("私钥信息错误")
	}
	return crypto.UnmarshalRsaPrivateKey(block.Bytes)
}

func (k *CryptoKeyInfo) MarshalPrivateKey(privateKey crypto.PrivKey) ([]byte, error) {
	return privateKey.Raw()
}

func (k *CryptoKeyInfo) MarshalRSAPrivateKeyToPem(privateKey crypto.PrivKey) ([]byte, error) {
	if privateKey.Type() != crypto.RSA {
		return nil, errors.New("公钥类型错误")
	}
	var privateBuff bytes.Buffer
	privateKeyBytes, _ := privateKey.Raw()
	//构建一个pem.Block结构体对象
	privateBlock := pem.Block{Type: "RSA Private Key", Bytes: privateKeyBytes}
	//将数据保存到文件
	_ = pem.Encode(&privateBuff, &privateBlock)
	return privateBuff.Bytes(), nil
}

func (k *CryptoKeyInfo) MarshaPublicKey(publicKey crypto.PubKey) ([]byte, error) {
	return publicKey.Raw()
}

func (k *CryptoKeyInfo) MarshaSelfPublicKey() ([]byte, error) {
	return k.GetPublicKey().Raw()
}

func (k *CryptoKeyInfo) UnMarshaPublicKey(publicKey []byte) (crypto.PubKey, error) {
	return crypto.UnmarshalRsaPublicKey(publicKey)
}

func (k *CryptoKeyInfo) MarshalRSAPublicKeyToPem(publicKey crypto.PubKey) ([]byte, error) {
	if publicKey.Type() != crypto.RSA {
		return nil, errors.New("公钥类型错误")
	}
	var privateBuff bytes.Buffer
	publicKeyBytes, _ := publicKey.Raw()
	//构建一个pem.Block结构体对象
	publicBlock := pem.Block{Type: "RSA Public Key", Bytes: publicKeyBytes}
	//将数据保存到文件
	_ = pem.Encode(&privateBuff, &publicBlock)
	return privateBuff.Bytes(), nil
}
