package contract

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"io/ioutil"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ContractClient define a interface para interagir com o contrato
type ContractClient interface {
	GetValue(ctx context.Context) (*big.Int, error)
	SetValue(ctx context.Context, value *big.Int, privateKey *ecdsa.PrivateKey) (common.Hash, error)
}

// SmartContract implementa ContractClient para o contrato SimpleStorage
type SmartContract struct {
	client          *ethclient.Client
	contractAddress common.Address
	parsedABI       abi.ABI
	chainID         *big.Int
}

// NewSmartContract cria uma nova instância de SmartContract
func NewSmartContract(nodeURL, abiPath, addressPath string) (*SmartContract, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, nodeURL)
	if err != nil {
		return nil, fmt.Errorf("erro conectando ao client Besu: %w", err)
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter Chain ID da rede: %w", err)
	}

	abiData, err := ioutil.ReadFile(abiPath)
	if err != nil {
		return nil, fmt.Errorf("erro lendo ABI do contrato: %w", err)
	}

	var artifact struct {
		ABI json.RawMessage `json:"abi"`
	}
	if err := json.Unmarshal(abiData, &artifact); err != nil {
		return nil, fmt.Errorf("erro parseando ABI do contrato: %w", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(string(artifact.ABI)))
	if err != nil {
		return nil, fmt.Errorf("erro convertendo ABI: %w", err)
	}

	addressData, err := ioutil.ReadFile(addressPath)
	if err != nil {
		return nil, fmt.Errorf("erro lendo endereço do contrato: %w", err)
	}

	var deploymentMap map[string]string
	if err := json.Unmarshal(addressData, &deploymentMap); err != nil {
		return nil, fmt.Errorf("erro lendo JSON de endereço do contrato: %w", err)
	}

	var contractAddress string
	for _, addr := range deploymentMap {
		contractAddress = addr
		break
	}

	return &SmartContract{
		client:          client,
		contractAddress: common.HexToAddress(contractAddress),
		parsedABI:       parsedABI,
		chainID:         chainID,
	}, nil
}

// GetValue busca o valor atual do contrato
func (sc *SmartContract) GetValue(ctx context.Context) (*big.Int, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	bound := bind.NewBoundContract(sc.contractAddress, sc.parsedABI, sc.client, sc.client, sc.client)

	var out []interface{}
	err := bound.Call(callOpts, &out, "get")
	if err != nil {
		return nil, fmt.Errorf("erro ao chamar função 'get' do contrato: %w", err)
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("retorno vazio da função 'get' do contrato")
	}

	val, ok := out[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("tipo de retorno inesperado da função 'get' do contrato: %T", out[0])
	}
	return val, nil
}

// SetValue define um novo valor no contrato
func (sc *SmartContract) SetValue(ctx context.Context, value *big.Int, privateKey *ecdsa.PrivateKey) (common.Hash, error) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Hash{}, fmt.Errorf("erro ao converter chave pública para ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := sc.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return common.Hash{}, fmt.Errorf("erro ao obter nonce da conta %s: %w", fromAddress.Hex(), err)
	}

	gasPrice, err := sc.client.SuggestGasPrice(ctx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("erro ao obter gas price: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, sc.chainID)
	if err != nil {
		return common.Hash{}, fmt.Errorf("erro ao criar transactor com ChainID %s: %w", sc.chainID.String(), err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(300000)
	auth.GasPrice = gasPrice

	bound := bind.NewBoundContract(sc.contractAddress, sc.parsedABI, sc.client, sc.client, sc.client)

	tx, err := bound.Transact(auth, "set", value)
	if err != nil {
		return common.Hash{}, fmt.Errorf("erro ao executar transação 'set' no contrato: %w", err)
	}

	return tx.Hash(), nil
}
