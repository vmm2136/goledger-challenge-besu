package config

import (
	"fmt"
	"os"
)

// Config armazena as configurações da aplicação
type Config struct {
	BesuNodeURL           string
	ContractABIPath       string
	ContractAddressesPath string
	ServerPort            string
	DatabaseURL           string
}

// LoadConfig carrega as configurações de variáveis de ambiente ou valores padrão
func LoadConfig() (*Config, error) {
	cfg := &Config{
		BesuNodeURL:           getEnvOrDefault("BESU_NODE_URL", "http://localhost:8545"),
		ContractABIPath:       getEnvOrDefault("CONTRACT_ABI_PATH", "../besu/artifacts/contracts/SimpleStorage.sol/SimpleStorage.json"),
		ContractAddressesPath: getEnvOrDefault("CONTRACT_ADDRESSES_PATH", "../besu/ignition/deployments/chain-1337/deployed_addresses.json"),
		ServerPort:            getEnvOrDefault("SERVER_PORT", "8080"),
		DatabaseURL:           getEnvOrDefault("DATABASE_URL", "root:root@tcp(127.0.0.1:3306)/besu_db?parseTime=true"),
	}

	if cfg.BesuNodeURL == "" {
		return nil, fmt.Errorf("BESU_NODE_URL não pode ser vazio")
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
