package ethutils

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"os"
)

// LoadPrivateKeyFromEnv carrega uma chave privada a partir de uma variável de ambiente
func LoadPrivateKeyFromEnv(envVarName string) (*ecdsa.PrivateKey, error) {
	keyHex := os.Getenv(envVarName)
	if keyHex == "" {
		return nil, fmt.Errorf("variável de ambiente %s para chave privada não encontrada ou vazia", envVarName)
	}

	privateKey, err := crypto.HexToECDSA(keyHex)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter chave privada de HEX: %w", err)
	}
	return privateKey, nil
}

// GetPublicKeyAddress retorna o endereço público de uma chave privada
func GetPublicKeyAddress(privateKey *ecdsa.PrivateKey) (common.Address, error) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, fmt.Errorf("erro ao converter chave pública para ECDSA")
	}
	return crypto.PubkeyToAddress(*publicKeyECDSA), nil
}
