package database

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"

	_ "github.com/lib/pq"
)

// DBClient define a interface para operações de banco de dados relacionadas ao contrato
type DBClient interface {
	GetContractValue(ctx context.Context, key string) (*big.Int, error)
	SaveContractValue(ctx context.Context, key string, value *big.Int) error
	ValidateContractValue(ctx context.Context, key string, expectedValue *big.Int) (bool, error)
}

// SQLDBClient é a implementação para bancos de dados SQL
type SQLDBClient struct {
	db *sql.DB
}

// NewSQLDBClient cria uma nova instância de SQLDBClient e abre a conexão com o DB
func NewSQLDBClient(databaseURL string) (*SQLDBClient, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir conexão com o DB: %w", err)
	}

	if err = db.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("erro ao conectar ao DB: %w", err)
	}

	fmt.Println("Conexão com o banco de dados estabelecida com sucesso.")
	return &SQLDBClient{db: db}, nil
}

// GetContractValue obtém o valor salvo do banco de dados para uma chave
func (c *SQLDBClient) GetContractValue(ctx context.Context, key string) (*big.Int, error) {
	var valueStr string

	query := `SELECT contract_value FROM contract_values WHERE contract_key = $1 LIMIT 1` // Removido ORDER BY, pois esperamos apenas um.
	err := c.db.QueryRowContext(ctx, query, key).Scan(&valueStr)
	if err == sql.ErrNoRows {
		return big.NewInt(0), nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar valor para chave '%s' no DB: %w", key, err)
	}

	value, ok := new(big.Int).SetString(valueStr, 10)
	if !ok {
		return nil, fmt.Errorf("erro ao converter valor '%s' do DB para big.Int para chave '%s'", valueStr, key)
	}
	return value, nil
}

// SaveContractValue faz um UPSERT do valor, caso exista, atualiza, senão, insere
func (c *SQLDBClient) SaveContractValue(ctx context.Context, key string, value *big.Int) error {
	upsertSQL := `
	INSERT INTO contract_values (contract_key, contract_value)
	VALUES ($1, $2)
	ON CONFLICT (contract_key) DO UPDATE
	SET contract_value = EXCLUDED.contract_value,
	    created_at = CURRENT_TIMESTAMP; 
	`

	_, err := c.db.ExecContext(ctx, upsertSQL, key, value.String())
	if err != nil {
		return fmt.Errorf("erro ao salvar/atualizar valor '%s' para chave '%s' no DB: %w", value.String(), key, err)
	}
	return nil
}

// ValidateContractValue compara um valor fornecido com o valor no banco de dados para uma chave fornecida
func (c *SQLDBClient) ValidateContractValue(ctx context.Context, key string, expectedValue *big.Int) (bool, error) {
	currentDBValue, err := c.GetContractValue(ctx, key)
	if err != nil {
		return false, fmt.Errorf("erro ao obter valor do DB para validação da chave '%s': %w", key, err)
	}

	if expectedValue.Cmp(currentDBValue) == 0 {
		return true, nil
	}
	return false, nil
}
