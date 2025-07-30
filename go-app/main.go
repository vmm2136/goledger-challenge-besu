package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/vmm2136/besu_challenge/go-app/internal/handler"
	"github.com/vmm2136/besu_challenge/go-app/internal/router"
	"net/http"
	"os"

	"github.com/vmm2136/besu_challenge/go-app/internal/config"
	"github.com/vmm2136/besu_challenge/go-app/internal/contract"
	"github.com/vmm2136/besu_challenge/go-app/internal/database"
	"github.com/vmm2136/besu_challenge/go-app/internal/pkg/ethutils"
	"github.com/vmm2136/besu_challenge/go-app/internal/service"
)

func main() {
	godotenv.Load() // Carrega .env se existir

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Erro ao carregar configurações: %v\n", err)
		os.Exit(1)
	}

	privateKey, err := ethutils.LoadPrivateKeyFromEnv("BESU_TRANSACTOR_PRIVATE_KEY")
	if err != nil {
		fmt.Printf("Erro ao carregar chave privada: %v\n", err)
		os.Exit(1)
	}

	pubAddress, err := ethutils.GetPublicKeyAddress(privateKey)
	if err != nil {
		fmt.Printf("Erro ao obter endereço público: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Transator usando o endereço: %s\n", pubAddress.Hex())

	// 1. Inicializar a camada de Contrato (interage com a blockchain)
	contractClient, err := contract.NewSmartContract(cfg.BesuNodeURL, cfg.ContractABIPath, cfg.ContractAddressesPath)
	if err != nil {
		fmt.Printf("Erro ao inicializar SmartContract: %v\n", err)
		os.Exit(1)
	}

	// 2. Inicializar a camada de Banco de Dados (interage com o DB SQL)
	// Lembre-se de instalar o driver Go para o seu DB (ex: github.com/go-sql-driver/mysql)
	dbClient, err := database.NewSQLDBClient(cfg.DatabaseURL)
	if err != nil {
		fmt.Printf("Erro ao inicializar cliente de banco de dados: %v\n", err)
		os.Exit(1)
	}

	// 3. Inicializar a camada de Serviço (contém a lógica de negócio, incluindo SYNC)
	// Agora ele recebe tanto o client do contrato quanto o client do DB.
	contractService, err := service.NewContractService(contractClient, dbClient, privateKey)
	if err != nil {
		fmt.Printf("Erro ao inicializar ContractService: %v\n", err)
		os.Exit(1)
	}

	// 4. Inicializar a camada de Handler (expõe endpoints HTTP)
	h := handler.NewHandler(contractService)

	// 5. Configurar o Router (mapeia URLs para handlers)
	router := router.NewRouter(h)

	// 6. Iniciar o Servidor HTTP
	fmt.Printf("Servidor iniciado na porta :%s\n", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, router); err != nil {
		fmt.Printf("Erro ao iniciar servidor: %v\n", err)
		os.Exit(1)
	}
}
