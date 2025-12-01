# Ethereum Indexer Configuration

This document describes all configuration options for the `config.json` file.

Source: `worker/ethereum/sai-ethereum-indexer/config/config.go`

## Configuration Structure

```json
{
  "common": { ... },
  "specific": { ... },
  "contracts": [ ... ]
}
```

---

## `common` - Server Settings

Framework-level server configuration.

Source: `internal/config-internal/config-internal.go`

| Field | Type | Description |
|-------|------|-------------|
| `common.http_server.enabled` | bool | Enable HTTP server for API endpoints |
| `common.http_server.host` | string | HTTP server bind address (e.g., "0.0.0.0") |
| `common.http_server.port` | string | HTTP server port (e.g., "8881") |
| `common.socket_server.enabled` | bool | Enable raw socket server |
| `common.socket_server.host` | string | Socket server bind address |
| `common.socket_server.port` | string | Socket server port |
| `common.web_socket.enabled` | bool | Enable WebSocket client connection |
| `common.web_socket.token` | string | Auth token for WebSocket connection |
| `common.web_socket.url` | string | WebSocket server URL to connect to |

### Example

```json
"common": {
  "http_server": {
    "enabled": true,
    "host": "0.0.0.0",
    "port": "8881"
  },
  "socket_server": {
    "enabled": false,
    "host": "0.0.0.0",
    "port": "8809"
  },
  "web_socket": {
    "enabled": false,
    "token": "auth_token",
    "url": "localhost:8820"
  }
}
```

---

## `specific` - Indexer Settings

Ethereum indexer-specific configuration.

Source: `config/config.go:17-26`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `geth_server` | string | - | Ethereum JSON-RPC endpoint URL |
| `start_block` | int | 0 | Block number to start indexing from |
| `operations` | []string | [] | List of operations to perform (e.g., "exportTokens") |
| `sleep` | int | 2 | Seconds to wait between polling for new blocks |
| `skipFailedTransactions` | bool | false | Skip transactions that failed on-chain |

### `specific.storage` - Storage Service Connection

| Field | Type | Description |
|-------|------|-------------|
| `collection` | string | MongoDB collection name for indexed data |
| `token` | string | Auth token for storage service |
| `url` | string | Storage service URL (sai-storage-mongo) |
| `email` | string | Email for auth (alternative to token) |
| `password` | string | Password for auth (alternative to token) |

### `specific.notifier` - Notification Service

| Field | Type | Description |
|-------|------|-------------|
| `sender_id` | string | Identifier for notification source |
| `token` | string | Auth token for notifier service |
| `url` | string | Notifier service URL |
| `email` | string | Email for auth (alternative to token) |
| `password` | string | Password for auth (alternative to token) |

### `specific.websocket` - WebSocket Output

| Field | Type | Description |
|-------|------|-------------|
| `token` | string | Auth token for WebSocket connection |
| `url` | string | WebSocket server URL to push updates to |

### Example

```json
"specific": {
  "geth_server": "https://data-seed-prebsc-1-s1.bnbchain.org:8545",
  "storage": {
    "collection": "Ethereum",
    "token": "12345",
    "url": "http://storage.local:8880",
    "email": "",
    "password": ""
  },
  "notifier": {
    "sender_id": "Ethereum",
    "token": "notification_token",
    "url": "http://localhost:8885",
    "email": "",
    "password": ""
  },
  "start_block": 42108003,
  "operations": ["exportTokens"],
  "skipFailedTransactions": true,
  "sleep": 20,
  "websocket": {
    "token": "ws_token",
    "url": "http://test-websocket:8820"
  }
}
```

---

## `contracts` - Smart Contract Definitions

Define contracts to monitor and index. Can also be stored in separate `contracts.json` file.

Source: `config/config.go:46-60`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | No | Human-readable contract name |
| `address` | string | Yes | Contract address on chain (0x...) |
| `abi` | string | Yes | Contract ABI as JSON string |
| `start_block` | int | Yes | Block number to start indexing this contract from |

### Example

```json
"contracts": [
  {
    "name": "Bridge",
    "address": "0x719CAe5e3d135364e5Ef5AAd386985D86A0E7813",
    "abi": "[{\"inputs\":[...]}]",
    "start_block": 42108003
  }
]
```

---

## Complete Example

```json
{
  "common": {
    "http_server": {
      "enabled": true,
      "host": "0.0.0.0",
      "port": "8881"
    },
    "socket_server": {
      "enabled": false,
      "host": "0.0.0.0",
      "port": "8809"
    },
    "web_socket": {
      "enabled": false,
      "token": "1",
      "url": "localhost:8820"
    }
  },
  "specific": {
    "geth_server": "https://data-seed-prebsc-1-s1.bnbchain.org:8545",
    "storage": {
      "collection": "Ethereum",
      "token": "12345",
      "url": "http://storage.local:8880",
      "email": "",
      "password": ""
    },
    "notifier": {
      "sender_id": "Ethereum",
      "token": "notification_token",
      "url": "http://localhost:8885",
      "email": "",
      "password": ""
    },
    "start_block": 42108003,
    "operations": ["exportTokens"],
    "skipFailedTransactions": true,
    "sleep": 20,
    "websocket": {
      "token": "ws_token",
      "url": "http://test-websocket:8820"
    }
  },
  "contracts": [
    {
      "name": "Bridge",
      "address": "0x719CAe5e3d135364e5Ef5AAd386985D86A0E7813",
      "abi": "[...]",
      "start_block": 42108003
    }
  ]
}
```

---

## Environment Variables

No environment variables are directly used by this service. All configuration is file-based.
