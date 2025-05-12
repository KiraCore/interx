# KIRA Interoperability Microservices Architecture

This repository contains a set of seven microservices designed to provide interoperability between blockchain networks, primarily focusing on Cosmos and Ethereum ecosystems.

## Architecture Overview

The system consists of the following microservices:

1. **Manager** - Main P2P load balancer and HTTP server for incoming requests
2. **Proxy** - Converts legacy interx paths to Manager requests
3. **Cosmos Indexer** - Indexes Cosmos blockchain data
4. **Cosmos Interaction** - Creates, signs, and publishes Cosmos transactions
5. **Ethereum Indexer** - Indexes Ethereum blockchain data
6. **Ethereum Interaction** - Creates, signs, and publishes Ethereum transactions
7. **Storage** - MongoDB storage for indexed blocks and transactions
   
## Install new node

1. Clone the project
2. Change ./manager/config.yml

```yaml
cosmos:
  node:
    json_rpc: "x.x.x.x:9090"            //Kira gRPC address
    tendermint: "http://x.x.x.x:26657"  //Rira tendermint address
p2p:
    id: "x"                     //Node sequence number
    address: "0.0.0.0:9000"     //Node bind address
    peers: ["x.x.x.x:9000"]     //Node peers (main node external IP:PORT)
    max_peers: 2                //Maximum node slots to accept connections
```
3. Execute: `docker compose up -d --build` in the root project directory
   
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
- `chain_id`: chain id from the config 
- 
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

## Getting Started

1. Configure the Manager service with appropriate settings.
2. Start all services using Docker Compose.
3. The system will be accessible through the Proxy service for legacy requests or directly through the Manager for new format requests.
