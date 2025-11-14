# Interx: KIRA Cross-Chain Microservices

Interx is a modular microservices architecture that provides cross-chain interoperability and indexing for the KIRA blockchain ecosystem. It serves as the gateway between blockchain networks, enabling seamless interaction with Cosmos (Sekai) and Ethereum-based chains through a distributed, load-balanced architecture.

## Part of KIRA Infrastructure

Interx is a core component of the [Sekin](https://github.com/KiraCore/sekin) infrastructure stack, working alongside:
- **Sekai**: KIRA's Cosmos SDK-based blockchain node
- **Shidai**: Orchestration and lifecycle management layer
- **Syslog-ng**: Centralized logging infrastructure

While Interx can be deployed standalone, it is designed to integrate seamlessly with the full KIRA stack for production deployments.

## Key Features

- **Multi-Chain Support**: Native integration with Cosmos SDK chains (KIRA/Sekai) and Ethereum/EVM networks
- **Distributed Load Balancing**: P2P-based request distribution across multiple nodes with intelligent routing
- **Comprehensive Indexing**: Real-time blockchain data indexing for blocks, transactions, and smart contract events
- **Transaction Management**: Create, sign, and broadcast transactions across supported chains
- **Legacy Compatibility**: Backward-compatible proxy layer for existing Interx API consumers
- **Horizontal Scaling**: Add nodes dynamically to handle increased load
- **Faucet Integration**: Built-in token faucet for testnet operations
- **Smart Contract Support**: Ethereum contract interaction and event monitoring

## Architecture Overview

Interx deploys seven specialized microservices that work together to provide comprehensive blockchain interoperability:

### Core Services

1. **Manager** (ports 8080, 9000/udp)
   - Main entry point and P2P load balancer
   - Distributes requests across the network
   - HTTP API server for incoming requests

2. **Proxy** (port 80 → 8080)
   - Legacy API compatibility layer
   - Converts traditional HTTP requests to Manager format
   - Maintains backward compatibility with older Interx deployments

### Blockchain Workers

3. **Cosmos Indexer** (port 8883)
   - Continuous indexing of Cosmos/KIRA blockchain data
   - Monitors blocks and transactions
   - Stores data via Storage service

4. **Cosmos Interaction** (port 8884)
   - Transaction creation and signing for Cosmos chains
   - Broadcast transaction management
   - Supports multiple broadcast modes (sync, async, block)

5. **Ethereum Indexer** (port 8881)
   - Indexes Ethereum and EVM-compatible chains
   - Smart contract event monitoring
   - Block and transaction data collection

6. **Ethereum Interaction** (port 8882)
   - Ethereum transaction creation and signing
   - Contract interaction handling
   - Multi-chain support via chain IDs

### Data Layer

7. **Storage** (port 8880)
   - MongoDB-based persistence layer
   - Stores indexed blockchain data
   - Accessed exclusively through Manager service

## Quick Start

### Standalone Deployment

1. **Clone the repository**
   ```bash
   git clone https://github.com/KiraCore/interx.git
   cd interx
   ```

2. **Create the external network**
   ```bash
   docker network create interx_default
   ```

3. **Configure Manager service**

   Edit `./manager/config.yml` with your node settings:
   ```yaml
   cosmos:
     node:
       json_rpc: "x.x.x.x:9090"            # KIRA gRPC address
       tendermint: "http://x.x.x.x:26657"  # KIRA Tendermint address

   p2p:
     id: "1"                     # Unique node identifier
     address: "0.0.0.0:9000"     # P2P bind address
     peers: ["x.x.x.x:9000"]     # List of peer nodes (empty for first node)
     max_peers: 2                # Maximum peer connections
   ```

4. **Launch all services**
   ```bash
   docker compose up -d --build
   ```

5. **Access the API**
   - Manager API: `http://localhost:8080`
   - Proxy (legacy): `http://localhost:80`

### Deployment with Sekin Stack

For production deployments, Interx is typically deployed as part of the [Sekin infrastructure stack](https://github.com/KiraCore/sekin). The Sekin stack provides:
- Pre-configured integration with Sekai nodes
- Centralized logging via syslog-ng
- Orchestration through Shidai
- Standardized networking (10.1.0.0/16)

Refer to the [Sekin documentation](https://github.com/KiraCore/sekin) for full stack deployment.

## P2P Load Balancing

Interx implements a distributed load balancing system through P2P communication between Manager nodes:

- **Metrics Collection**: Each node tracks CPU load, memory usage, and requests per second over a configurable window
- **Intelligent Routing**: Requests are automatically forwarded to less-loaded peers when local resources are constrained
- **Threshold-Based**: Load balancing decisions use configurable thresholds to prevent unnecessary routing
- **Peer Discovery**: Nodes maintain connections with configured peers and can accept new connections up to `max_peers`

This architecture allows horizontal scaling by adding more Interx nodes to the P2P network.

## Service Details

### Manager

The Manager service acts as the central entry point for all API requests and distributes the load across the network using P2P communication. It accepts requests in the following format:

```json
{
  "method": "ethereum || cosmos || rosetta || bitcoin || any",
  "data": {
    "method": "POST || GET",
    "path": "Endpoint",
    "payload": {}
  }
}
```

- `method`: Indicates which blockchain handler to use (`any` is for non-chain-specific endpoints, aggregation, etc.)
- `data.method`: HTTP method (GET/POST)
- `data.path`: Endpoint path (for Cosmos) or RPC method (for Ethereum)
- `data.payload`: Request body for POST or query parameters for GET

### Proxy

The Proxy service translates traditional HTTP requests into the Manager format. It supports both GET and POST requests following the legacy interx system structure:

**Example Cosmos|Kira request:**
```
GET /cosmos/bank/v1beta1/supply
```
or
```
GET /kira/gov/all_roles
```
**Example Ethereum request:**
```
GET /ethereum/{chain_id}/eth_blockNumber
```
Where `chain_id` corresponds to the chain identifier configured in the Manager config.


### Cosmos Indexer

This service continuously indexes Cosmos blockchain data (blocks and transactions) and stores them in the MongoDB database. It operates independently and doesn't accept external calls.

### Cosmos Interaction

Responsible for creating, signing, and publishing Cosmos transactions. In this architecture, it only accepts calls from the Manager service.

### Ethereum Indexer

Similar to the Cosmos Indexer, this service indexes Ethereum blockchain data and stores it in the MongoDB database. It also operates independently without accepting external calls.

### Ethereum Interaction

Handles the creation, signing, and publishing of Ethereum transactions. Like the Cosmos Interaction service, it only accepts calls from the Manager.

### Storage

MongoDB-based storage service for indexed blockchain data. In this architecture, it only accepts calls from the Manager service.

## API Usage Examples

### Using the Proxy (Legacy Format)

The Proxy service provides backward-compatible HTTP endpoints:

```bash
# Query Cosmos/KIRA chain
curl http://localhost/cosmos/bank/v1beta1/supply
curl http://localhost/kira/gov/all_roles

# Query Ethereum chain
curl http://localhost/ethereum/1/eth_blockNumber
curl http://localhost/ethereum/56/eth_getBalance?address=0x...
```

### Using the Manager (New Format)

Direct Manager API requests use a structured JSON format:

```bash
# Cosmos query
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "method": "cosmos",
    "data": {
      "method": "GET",
      "path": "/cosmos/bank/v1beta1/supply",
      "payload": {}
    }
  }'

# Ethereum query
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "method": "ethereum",
    "data": {
      "method": "POST",
      "path": "eth_blockNumber",
      "payload": {}
    }
  }'
```

## Configuration

The main configuration is done in the Manager service:

```yaml
storage:
  token: ""  # Access token for the storage service
  url: "http://worker-sai-storage:8880"  # Docker container address (do not change)

ethereum:
  interaction: "http://worker-sai-ethereum-interaction:8882" # Docker container address (do not change)
  nodes:
    chain1: "https://data-seed-prebsc-1-s1.bnbchain.org:8545" #chain node address, chain1 -> chain_id
  token: ""  # Access token for the Ethereum interaction service
  retries: 1  # Number of request retries
  retry_delay: 10  # Delay before retry (seconds)
  rate_limit: 10  # Maximum allowed requests per second

cosmos:
  node:
    json_rpc: "157.180.16.117:9090" # Cosmos gRPC address
    tendermint: "http://157.180.16.117:26657" #Cosmos tendermint address
  tx_modes: # Accepted transaction broadcast types
    sync: true
    async: true
    block: true
  faucet: # Faucet configuration
    faucet_amounts:
      ukex: 1000
    faucet_minimum_amounts:
      ukex: 100
    fee_amounts:
      ukex: 100
    time_limit: 3600
  gw_timeout: 30 # Gateway request timeout
  interaction: "http://worker-sai-cosmos-interaction:8884"  # Docker container address (do not change)
  token: ""  # Access token for the Cosmos interaction service
  retries: 1  # Number of request retries
  retry_delay: 10  # Delay before retry (seconds)
  rate_limit: 10  # Maximum allowed requests per second

p2p:
  id: "1"  # UNIQUE identifier for this node
  address: "127.0.0.1:9000"  # Address where P2P listens for incoming connections
  peers: []  # List of initial peers (empty means this is the first node in the network)
  max_peers: 2  # Maximum number of connections accepted by the server

balancer:
  window_size: 60  # Interval in seconds for metrics collection (CPU load, memory usage, RPS)
  threshold: 0.2  # Threshold for load balancing decisions
```

### Load Balancing Logic

The balancer makes decisions based on system metrics collected over the configured window size:

- If the minimum overall load among all nodes is less than the current node's load by more than the threshold value, the request is forwarded to the less loaded node.
- Metrics include CPU load, memory usage, and requests per second.

## Development

### Building from Source

Each microservice can be built independently using its respective Dockerfile:

```bash
# Build a specific worker
cd worker/cosmos/sai-cosmos-indexer
docker build -t interx-cosmos-indexer .

# Build the manager
cd manager
docker build -t interx-manager .

# Build all services via docker-compose
docker compose build
```

### Project Structure

```
interx/
├── manager/              # Main API and P2P load balancer
├── proxy/                # Legacy API compatibility layer
├── worker/
│   ├── cosmos/
│   │   ├── sai-cosmos-indexer/       # Cosmos blockchain indexer
│   │   └── sai-cosmos-interaction/   # Cosmos transaction handler
│   ├── ethereum/
│   │   ├── sai-ethereum-indexer/           # Ethereum blockchain indexer
│   │   └── sai-ethereum-contract-interaction/ # Ethereum transaction handler
│   └── sai-storage-mongo/            # MongoDB storage service
└── docker-compose.yml    # Orchestration configuration
```

### Configuration Files

Each service requires a `config.yml` file mounted at `/srv/config.yml`:
- `manager/config.yml` - Main configuration (see Configuration section)
- `proxy/config.yml` - Proxy service settings
- `worker/*/config.yml` - Individual worker configurations

## Additional Resources

- [Sekin Stack](https://github.com/KiraCore/sekin) - Full KIRA infrastructure deployment
- [KIRA Network](https://kira.network) - Official KIRA website
- [Documentation](https://docs.kira.network) - Comprehensive KIRA documentation

## License

This project is part of the KIRA Network infrastructure.
