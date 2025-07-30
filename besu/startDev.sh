#! /bin/bash

echo "Starting QBFT Besu local network"

# Check if besu binary is installed
echo "Checking if besu is installed..."
if ! [ -x "$(command -v besu)" ]; then
  echo "Error: besu is not installed. Go to https://besu.hyperledger.org/private-networks/get-started/install/binary-distribution" >&2
  exit 1
fi

# Check if Hardhat is installed
echo "Checking if Hardhat is installed..."
if ! npx --no-install hardhat > /dev/null 2>&1; then
  echo "Error: Hardhat is not installed. Run: npx hardhat" >&2
  exit 1
fi


echo "Cleaning up previous data"

# Clean up containers
docker rm -f besu-node-0 besu-node-1 besu-node-2 besu-node-3 

# Clean up data dir from each node
rm -rf node/besu-0/data
rm -rf node/besu-1/data
rm -rf node/besu-2/data
rm -rf node/besu-3/data

rm -rf genesis
rm -rf ignition/deployments
rm -rf artifacts
rm -rf cache

rm -rf _tmp

# Recreate data dir for each node
mkdir -p node/besu-0/data
mkdir -p node/besu-1/data
mkdir -p node/besu-2/data
mkdir -p node/besu-3/data

# generate keys and genesis
mkdir _tmp && cd _tmp
besu operator generate-blockchain-config --config-file=../config/qbftConfigFile.json --to=networkFiles --private-key-file-name=key

cd ..

counter=0  # Initialize the counter
# Copy keys to each node
for folder in _tmp/networkFiles/keys/*; do
  folderName=$(basename "$folder")
  cp _tmp/networkFiles/keys/$folderName/key node/besu-$counter/data/key
  cp _tmp/networkFiles/keys/$folderName/key.pub node/besu-$counter/data/key.pub
  ((counter++))
done

# Copy genesis to each node
mkdir genesis && cp _tmp/networkFiles/genesis.json genesis/genesis.json

jq '.alloc += {
  "fe3b557e8fb62b89f4916b721be55ceb828dbd73": {
    "privateKey": "8f2a55949038a9610f50fb23b5883af3b4ecb3c3bb792cbcefbd1542c692be63",
    "comment": "This is Alice'\''s Key. Private key and this comment are ignored. In a real chain, the private key should NOT be stored",
    "balance": "80000000000000000000000"
  },
  "627306090abaB3A6e1400e9345bC60c78a8BEf57": {
    "privateKey": "c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3",
    "comment": "This is the contract'\''s Key. Private key and this comment are ignored. In a real chain, the private key should NOT be stored",
    "balance": "70000000000000000000000"
  },
  "f17f52151EbEF6C7334FAD080c5704D77216b732": {
    "privateKey": "ae6ae8e5ccbfb04590405997ee2d52d2b330726137b875053c36d94e974d162f",
    "comment": "This is Bob'\''s Key. Private key and this comment are ignored. In a real chain, the private key should NOT be stored",
    "balance": "90000000000000000000000"
  }
}' genesis/genesis.json > temp.json && mv temp.json genesis/genesis.json

rm -rf _tmp

if ! docker network ls | grep -q besu_network; then
  docker network create besu_network
fi

echo "Starting Postgres for value storage..."
docker-compose -f docker/docker-compose-postgres.yaml up -d

# Espera o banco ficar disponível
echo "Waiting for Postgres to be ready..."
until docker exec besu_postgres pg_isready -U besu > /dev/null 2>&1; do
  sleep 1
done

echo "Starting bootnode"
docker-compose -f docker/docker-compose-bootnode.yaml up -d

# Retrieve bootnode enode address
max_retries=30
retry_delay=1
retry_count=0

while [ $retry_count -lt $max_retries ]; do
  export ENODE=$(curl -X POST --data '{"jsonrpc":"2.0","method":"net_enode","params":[],"id":1}' http://127.0.0.1:8545 | jq -r '.result')

  if [ -n "$ENODE" ]; then
    if [ "$ENODE" != "null" ]; then
      echo "ENODE retrieved successfully."
      break
    fi
  else
    echo "Failed to retrieve ENODE. Retrying in $retry_delay seconds..."
    sleep $retry_delay
    ((retry_count++))
  fi
done

if [ $retry_count -eq $max_retries ]; then
  echo "Max retries reached. Unable to retrieve ENODE."
fi

echo "ENODE: $ENODE"

export E_ADDRESS="${ENODE#enode://}"
export DOCKER_NODE_1_ADDRESS=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' besu-node-0)
export E_ADDRESS=$(echo $E_ADDRESS | sed -e "s/127.0.0.1/$DOCKER_NODE_1_ADDRESS/g")
echo $E_ADDRESS

sed "s/<ENODE>/enode:\/\/$E_ADDRESS/g" docker/templates/docker-compose-nodes.yaml > docker/docker-compose-nodes.yaml

echo "Starting nodes"
docker-compose -f docker/docker-compose-nodes.yaml up -d

echo "============================="
echo "Network started successfully!"
echo "============================="

echo "Setting up Hardhat for contract deployment..."

cd contracts

# Install dependencies if not already installed
npm install --save-dev hardhat ethers @openzeppelin/contracts

echo "Compiling contracts..."
npx hardhat compile

cd ..

echo "Deploying contract..."
npx hardhat ignition deploy ./ignition/modules/deploy.js --network besu << EOF
y
EOF