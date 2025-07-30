# API em Go com Besu e PostgreSQL

Este projeto implementa uma API REST em Go para interagir com um contrato inteligente em uma rede Besu e persistir seu estado em um banco de dados PostgreSQL.

---

## 🚀 Visão Geral e Objetivo

A aplicação demonstra a **interação de uma API Go com tecnologias blockchain (Besu) e SQL (PostgreSQL)**. Ela permite:
* **Definir (SET)** e **Recuperar (GET)** valores do contrato na blockchain.
* **Sincronizar (SYNC)** o valor da blockchain para o PostgreSQL.
* **Verificar (CHECK)** a consistência entre o valor da blockchain e o valor no banco de dados.

O projeto foi construído com foco em **arquitetura limpa**, **boas práticas de Go** e princípios **SOLID**, promovendo **código testável e de fácil manutenção**.

---

## 🏛️ Arquitetura da Aplicação

A aplicação segue uma **arquitetura em camadas** com clara separação de responsabilidades.

**Destaques de Design:**
* **Interfaces e Injeção de Dependência:** Utilizadas extensivamente para desacoplar as camadas, facilitando a testabilidade e a flexibilidade.
* **UPSERT no DB:** O valor do contrato é armazenado de forma única por uma `contract_key`, garantindo que o banco de dados sempre reflita o **estado atual mais recente**, sem criar registros duplicados a cada sincronização.

---

## 🛠️ Configuração e Como Rodar

Este projeto utiliza **Docker** e **Docker Compose** para orquestrar o ambiente de desenvolvimento (rede Besu, PostgreSQL e a aplicação Go).

### Pré-requisitos:
- NPM and NPX (https://www.npmjs.com/get-npm)
- Hardhat (https://hardhat.org/getting-started/)
- Docker and Docker Compose (https://www.docker.com/)
- Besu (https://besu.hyperledger.org/private-networks/get-started/install/binary-distribution)
- Go (https://golang.org/dl/)
- Insomnia/Postman (para testar a API)

### Passos:

1.  **Clone o repositório:**
    ```bash
    git clone <https://github.com/vmm2136/goledger-challenge-besu>
    cd <goledger-challenge-besu>
    ```

2.  **Crie e configure o arquivo `.env`** na pasta go-app:
   
    ❗❗❗ ATENÇÃO ❗❗❗ Os valores abaixo estão explicitamente descritos apenas para facilitar o teste desta aplicação, não refletindo nenhuma conexão real (produção).
    
    ```env
    BESU_TRANSACTOR_PRIVATE_KEY="4f3edf983ac636a65a842ce7c78d9aa706d3b113b2c213a1f1f0eb46e5b21678"
    BESU_NODE_URL="http://localhost:8545"
    CONTRACT_ABI_PATH="../besu/artifacts/contracts/SimpleStorage.sol/SimpleStorage.json"
    CONTRACT_ADDRESSES_PATH="../besu/ignition/deployments/chain-1337/deployed_addresses.json"
    SERVER_PORT="8080"
    DATABASE_URL="postgres://besu:besu123@localhost:5433/besu?sslmode=disable"
    ```

4.  **Inicie o ambiente com o script de desenvolvimento:**
    ```bash
    ./startDev.sh
    ```
    Este script automatiza a instalação de dependências Hardhat, compilação/deploy do contrato, inicialização da rede Besu e, **via `docker-compose-postgres.yaml`, sobe o PostgreSQL (que automaticamente cria a tabela `contract_values`).**

---

## ⚡ Testando a API

Com a aplicação rodando (em `http://localhost:8080`), utilize sua ferramenta preferida (Insomnia/Postman) para interagir com os endpoints:

* **`GET /value`**: Recupera o valor atual do contrato na **blockchain**.
* **`POST /value`**: Define um novo valor no contrato na **blockchain**.
    * **Body:** `{"value": <número inteiro>}`
* **`POST /sync`**: Sincroniza o valor da **blockchain** para o **PostgreSQL**.
* **`GET /check`**: Compara o valor da **blockchain** com o valor no **PostgreSQL**. Retorna `true` se iguais, `false` caso contrário.

---

## 💡 Considerações Adicionais

* **Automação do Ambiente:** A inclusão do `docker-compose-postgres.yaml` e a atualização do `startDev.sh` foram implementadas para garantir um **ambiente de desenvolvimento completo e de fácil reprodução**, englobando Besu e PostgreSQL. O script cuida do deploy do contrato e da criação da tabela no banco.
* **Gerenciamento de Segredos:** A chave privada do transator é carregada via variável de ambiente, evitando sua exposição no código fonte.
* **Tratamento de Edge Cases:** A lógica de leitura do DB retorna `0` quando uma `contract_key` não é encontrada (em vez de erro), permitindo que as funções de `SYNC` e `CHECK` operem de forma fluida mesmo no estado inicial do banco.

---
