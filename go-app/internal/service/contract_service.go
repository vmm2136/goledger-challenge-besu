package service

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vmm2136/besu_challenge/go-app/internal/contract"
	"github.com/vmm2136/besu_challenge/go-app/internal/database"
	"github.com/vmm2136/besu_challenge/go-app/internal/pkg/ethutils"
)

// Constante que será  usada no banco como chave do contrato
const SimpleStorageValueKey = "simple_storage_current_value"

// ContractService define a interface para a lógica do contrato
type ContractService interface {
	GetCurrentValue(ctx context.Context) (*big.Int, common.Address, error)
	SetNewValue(ctx context.Context, value int64) (common.Hash, error)
	SyncContractValue(ctx context.Context) (networkValue *big.Int, dbValue *big.Int, err error)
	CheckContractValue(ctx context.Context) (bool, *big.Int, *big.Int, error)
}

// contractServiceImpl implementa ContractService
type contractServiceImpl struct {
	contractClient contract.ContractClient
	dbClient       database.DBClient
	privateKey     *ecdsa.PrivateKey
}

// NewContractService cria uma nova instância de ContractService
func NewContractService(client contract.ContractClient, dbClient database.DBClient, privateKey *ecdsa.PrivateKey) (ContractService, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("chave privada para o serviço não pode ser nula")
	}
	if dbClient == nil {
		return nil, fmt.Errorf("cliente de banco de dados não pode ser nulo")
	}
	return &contractServiceImpl{
		contractClient: client,
		dbClient:       dbClient,
		privateKey:     privateKey,
	}, nil
}

// GetCurrentValue obtém o valor atual do contrato e o endereço do transator
func (s *contractServiceImpl) GetCurrentValue(ctx context.Context) (*big.Int, common.Address, error) {
	value, err := s.contractClient.GetValue(ctx) // Obtém da rede
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("erro ao obter valor do contrato: %w", err)
	}

	transactorAddress, err := ethutils.GetPublicKeyAddress(s.privateKey)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("erro ao obter endereço público do transator: %w", err)
	}

	return value, transactorAddress, nil
}

// SetNewValue define um novo valor no contrato
func (s *contractServiceImpl) SetNewValue(ctx context.Context, value int64) (common.Hash, error) {
	txHash, err := s.contractClient.SetValue(ctx, big.NewInt(value), s.privateKey)
	if err != nil {
		return common.Hash{}, fmt.Errorf("erro ao definir novo valor no contrato: %w", err)
	}

	return txHash, nil
}

// SyncContractValue busca o valor do contrato na rede e o sincroniza com o banco de dados
func (s *contractServiceImpl) SyncContractValue(ctx context.Context) (*big.Int, *big.Int, error) {
	networkValue, err := s.contractClient.GetValue(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao obter valor da rede para sincronização: %w", err)
	}

	dbValue, err := s.dbClient.GetContractValue(ctx, SimpleStorageValueKey)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao obter valor do DB para sincronização (chave %s): %w", SimpleStorageValueKey, err)
	}

	if networkValue.Cmp(dbValue) != 0 {
		fmt.Printf("Sincronizando: Valor na rede (%s) difere do valor no DB (%s) para chave '%s'. Atualizando DB...\n",
			networkValue.String(), dbValue.String(), SimpleStorageValueKey)
		err = s.dbClient.SaveContractValue(ctx, SimpleStorageValueKey, networkValue)
		if err != nil {
			return nil, nil, fmt.Errorf("erro ao salvar novo valor no DB durante sincronização (chave %s): %w", SimpleStorageValueKey, err)
		}
		dbValue = networkValue
	} else {
		fmt.Printf("Sincronização: Valores da rede (%s) e DB (%s) já são iguais para chave '%s'.\n",
			networkValue.String(), dbValue.String(), SimpleStorageValueKey)
	}

	return networkValue, dbValue, nil
}

// CheckContractValue busca o valor do contrato na rede e o compara com o valor no banco de dados
func (s *contractServiceImpl) CheckContractValue(ctx context.Context) (bool, *big.Int, *big.Int, error) {
	networkValue, err := s.contractClient.GetValue(ctx)
	if err != nil {
		return false, nil, nil, fmt.Errorf("erro ao obter valor da rede para verificação: %w", err)
	}

	dbValue, err := s.dbClient.GetContractValue(ctx, SimpleStorageValueKey)
	if err != nil {
		return false, nil, nil, fmt.Errorf("erro ao obter valor do DB para verificação (chave %s): %w", SimpleStorageValueKey, err)
	}

	areEqual := networkValue.Cmp(dbValue) == 0

	return areEqual, networkValue, dbValue, nil
}
