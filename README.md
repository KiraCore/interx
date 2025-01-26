# INTERX

INTERX is an interchain engine, proxy, load balancer & security gateway service for communication between backend and frontend.  
It connects to the node using the GRPC endpoint as well as the RPC endpoint ([`Tendermint RPC`](https://docs.tendermint.com/master/rpc/)).

## Releases

All files in the [KIRA releases](https://github.com/KiraCore/interx/releases) are always signed with [cosign](https://github.com/sigstore/cosign/releases) and listed with a corresponding `SHA256` checksum.  
You should NEVER install anything on your machine unless you have verified the integrity of the files!

To learn more about how to install `cosign` and verify the integrity of files, click [here](./cosign.md).

## Setup

### Prerequisites

Essential dependencies and tools required to build INTERX can be found [here](./PREREQUISITES.md).  
Make sure you run all prerequisite commands successfully before you continue with local development or further installation.

### Installation from Github

The binary (`interxd`) will be installed in the `$GOBIN` directory. Please make sure to add it to your system `PATH`.

```bash
cd $HOME && rm -fvr ./interx && INTERX_BRANCH="<branch-name-or-hash>" && \
 git clone https://github.com/KiraCore/interx.git -b $INTERX_BRANCH && \
 cd ./interx && chmod -R 777 ./scripts && \
 make install && echo "SUCCESS" || echo "FAILED"
```

### How to start

Simple start:

```bash
interxd init
interxd start
```

Generates the configuration file.

Parameters:

- **`config`** - The interx configuration file path. (default = `./config.json`)
- **`serve_https`** - Use HTTPS or HTTP. (default = `false`)
- **`grpc`** - The GRPC endpoint of the `sekaid`. (default = `dns:///0.0.0.0:9090`)
- **`rpc`** - The Tendermint RPC endpoint of the `sekaid`. (default = `http://0.0.0.0:26657`)
- **`port`** - The interx port. (default = `11000`)
- **`signing_mnemonic`** - The mnemonic file path or word seeds for the interx signing service. (default = auto-generated word seeds)
- **`status_sync`** - The time interval in seconds for INTERX to sync node status. (default = `5`)
- **`cache_dir`** - The interx cache directory path. (default = `cache`)
- **`max_cache_size`** - The maximum cache size. (default = `2GB`)
- **`caching_duration`** - The caching clear duration in seconds. (default = `5`)
- **`download_file_size_limitation`** - The maximum file size allowed for downloads. (default = `10MB`)
- **`faucet_mnemonic`** - The mnemonic file path or word seeds for the faucet service. (default = auto-generated word seeds)
- **`faucet_time_limit`** - The claim time limitation in seconds. (default = `20`)

`interxd start`

Start interx service.

Parameters:

- `config` - The interx configuration file path. (default = "./config.json")

### Configuration

Configurations are available using `config.json`.

### mnemonic

The 24-word seed string or the file path to the mnemonic file.
</b>
This is for generating interx private/public keys which will be used for response signing.

```bash
"mnemonic": "swap exercise equip shoot mad inside floor wheel loan visual stereo build frozen always bulb naive subway foster marine erosion shuffle flee action there"

"mnemonic": "interx.mnemonic"
```

### Cache

Cache configurations.

#### `status_sync`

Interx has a feature to sync node status.
`status_sync` refers the time in seconds and INTERX syncs node status every `status_sync` seconds.

```json
"status_sync": 5
```

#### `cache_dir`

Interx has a caching feature.
`cache_dir` refers the cache directory.

```json
"cache_dir": "cache"
```

#### `max_cache_size`

Interx has a gabage collection feature.
`max_cache_size` refers the maximum cache size. If cache size is over maximum cache size, it will remove random caches. (it remains 90% of maximum cache size)

```json
"max_cache_size": "2 GB"
```

#### `caching_duration`

Interx has a gabage collection feature.
`caching_duration` refers the caching clear duration in seconds

```json
"caching_duration": 10
```

#### `download_file_size_limitation`

Interx has a download feature.
`download_file_size_limitation` refers the maximum download file size.

```json
"download_file_size_limitation": "10 MB"
```

#### faucet

Interx has a faucet feature.

#### faucet_mnemonic

`mnemonic` refers the 24-words seed string or the file path to the mnemonic file.
It will be used to generate faucet account priv/pub keys and address.

```text
"mnemonic": "equip exercise shoot mad inside floor wheel loan visual stereo build frozen potato always bulb naive subway foster marine erosion shuffle flee action there"

"mnemonic": "faucet.mnemonic"
```

#### faucet_amounts

`faucet_amounts` refers the faucet amount for each tokens.

```json
"faucet_amounts": {
    "stake": 100000,
    "validatortoken": 100000,
    "ukex": 100000
},
```

#### faucet_minimum_amounts

`faucet_minimum_amounts` refers the faucet minimum amount for each tokens.

```json
"faucet_minimum_amounts": {
    "stake": 100,
    "validatortoken": 100,
    "ukex": 100
},
```

#### fee_amounts

`fee_amounts` refers the fee amount for faucet feature.
For `stake` token faucet, we can use different coins for fee. E.g. we can use `ukex` for `stake` token faucet.

```json
"fee_amounts": {
    "stake": "1000ukex",
    "validatortoken": "1000ukex",
    "ukex": "1000ukex"
}
```

#### time_limit

`time_limit` refers the claim time limitation in seconds.
Users can re-request faucet after `time_limit` seconds.

```json
"time_limit": 20
```

### rpc

`rpc` refers the RPC endpoint configurations.

### API

`API` refers the configurations for each endpoint.

#### GET

`GET` refers the configurations for each `GET` endpoint.

```json
"GET": {
    "/api/cosmos/status": {
        "rate_limit": 0.1,
        "auth_rate_limit": 1,
        "caching_duration": 30
    },
    "/api/cosmos/bank/supply": {
        "disable": true,
        "auth_rate_limit": 1,
        "caching_disable": true,
        "caching_duration": 30
    }
},
```

#### disable

`disable` refers the options to disable the endpoint.

```json
"disable": true
```

### rate_limit

`rate_limit` refers the rate limit for each endpoint.

```json
"rate_limit": 0.1
```

### auth_rate_limit

`auth_rate_limit` refers the auth rate limit for each endpoint.

```json
"auth_rate_limit": 1
```

### caching_disable

`caching_disable` refers the option to disable caching feature for each endpoint.

```json
"caching_disable": true
```

### caching_duration

`caching_duration` refers the customized caching duration time in seconds for each endpoint.

```json
"caching_duration": 30
```

### POST

`POST` refers the configurations for each `POST` endpoint.

This is the same as `GET` configurations.

```json
"POST": {
    "/api/cosmos/txs": {
        "disable": false,
        "rate_limit": 0.1
    }
}
```

## Networking

### Communication Between Sekai and INTERX

INTERX connects to sekai using following ports.

- `9090`: Connect to the GRPC endpoint for the node.
- `26657`: Connect to the Tendermint RPC endpoint for the node.

### Communication Between INTERX and Frontend

- INTERX uses `http` protocol by default.
It's possible to config `http` setting using `SERVE_HTTP`(default = true) env variable.

- INTERX uses `11000` port by default.
It's possible to config `port` setting using `PORT`(deafult = 11000) env variable.

## Communication

### /api/kira/metadata

Query functions metadata for `sekai`.

### /api/interx/metadata

Query functions metadata for `interx`.

### /api/rpc_methods

Query available RPC methods `interx` provides.

### /api/status

QueryStatus is a function to query the node status.

### /api/cosmos/auth/accounts/{address}

QueryAccount is a function to query the account info.

#### Parameters

- `address`: (`string`) The account address.

#### Example

GET <http://0.0.0.0:11000/api/cosmos/auth/accounts/kira1gaadckc6g8ne62dzmscgyqkx3sd5p26wrapekd>

#### /api/cosmos/bank/balances/{address}

QueryBalance is a function to query the account balances.

1- Parameters

- `address`: (`string`) The account address.

2- Example

GET <http://0.0.0.0:11000/api/cosmos/bank/balances/kira1gaadckc6g8ne62dzmscgyqkx3sd5p26wrapekd>

### `/api/cosmos/txs/{hash}`

QueryTransactionHash is a function to query transaction details from transaction hash.

1- Parameters

- `hash`: (`string`) The transaction hash. (e.g. 0x20.....)

2- Example

GET <http://0.0.0.0:11000/api/cosmos/txs/0x4A41257AC228F6CE476E9C9AD67BB98057412A22B035E1C0A4CCEB0E4E8E364D>

### `/api/kira/faucet`

Faucet is a function to claim tokens to the account for free. Returns the available faucet amount when 'claim' and 'token' is unset.

1- Parameters

- `claim`: (`string`, `optional`) The claim address.
- `token`: (`string`, `optional`) The claim token.

2- Example

GET <http://0.0.0.0:11000/api/kira/faucet>
GET <http://0.0.0.0:11000/api/kira/faucet?claim=kira1kdnep4lm3z6yd3pah0rzfu3dvudgwfjejs9ans&token=stake>

### `/api/withdraws`

Withdraws is a function to query withdraw transactions of the account.

1- Parameters

- `account`: (`string`) The Kira account address.
- `type`: (`string`, `optional`) The transaction type.
- `max`: (`int`, `optional`) The maximum number of the results. (1 ~ 1000)
- `last`: (`string`, `optional`) The last transaction hash.

2- Example

GET <http://0.0.0.0:11000/api/withdraws?account=kira1eyvuhkj9r28sutr6n5vxgckejz2qy3hvanjk7k&type=send&max=4>

### `/api/deposits`

Deposits is a function to query deposit transactions of the account.

1- Parameters

- `account`: (`string`) The Kira account address.
- `type`: (`string`, `optional`) The transaction type.
- `max`: (`int`, `optional`) The maximum number of the results. (1 ~ 1000)
- `last`: (`string`, `optional`) The last transaction hash.

2- Example

GET <http://0.0.0.0:11000/api/deposits?account=kira1h9s2k2s9624kdghp5ztcdgnausg77rdj9cyat6&type=Send>

### `/api/kira/gov/data_keys`

QueryDataReferenceKeys is a function to query data reference keys with pagination.

1- Parameters

- `limit`: (`number`) The limit of the query results. (like page size)
- `offset`: (`number`) The offset of the query results. (like page number)
- `count_total`: (`bool`, `optional`) The option to return total count of data reference keys.

2- Example

GET <http://0.0.0.0:11000/api/kira/gov/data_keys?limit=2&offset=0&count_total=true>

### `/api/kira/gov/data/{key}`

QueryDataReference is a function to query data reference by a key.

1- Parameters

- `key`: (`string`) The data reference key.

2- Example

GET <http://0.0.0.0:11000/api/kira/gov/data/data_reference_key>

### `/download/{module}/{key}`

Download is a function to download a data reference or arbitrary data.

1- Parameters

- `module`: (`string`) The module name. (e.g. DRR for data reference registry.)
- `key`: (`string`) The reference key. (It saves reference data with hashed name. e.g. 2CEE6B1689EDDDD6F08EB1EAEC7D3C4E.)

2- Example

GET <http://0.0.0.0:11000/download/DRR/2CEE6B1689EDDDD6F08EB1EAEC7D3C4E>

### `/api/cosmos/txs`

Broadcast is a function to broadcast signed transaction.

1- Parameters

- `tx`: (`object`) The signed transaction to be broadcasted.
- `mode`: (`string`) The transaction broadcast mode. (block, sync, async)

2- Example

POST <http://0.0.0.0:11000/api/cosmos/txs>

```json
{
 "tx": {
  "body": {
   "messages": [
    {
     "@type": "/cosmos.bank.v1beta1.MsgSend",
     "from_address": "kira1rsgnkecqgq575ynn6rczd96kc23uwtruqx9m0m",
     "to_address": "kira1h9s2k2s9624kdghp5ztcdgnausg77rdj9cyat6",
     "amount": [
      {
       "denom": "ukex",
       "amount": "250"
      }
     ]
    }
   ],
   "memo": "",
   "timeout_height": "0",
   "extension_options": [],
   "non_critical_extension_options": []
  },
  "auth_info": {
   "signer_infos": [
    {
     "public_key": {
      "secp256k1": "Alm0A4BIQyUWy8KXjP1BRMePguZWgFKQa5hzzwRlu3I8"
     },
     "mode_info": {
      "single": {
       "mode": "SIGN_MODE_DIRECT"
      }
     },
     "sequence": "0"
    }
   ],
   "fee": {
    "amount": [],
    "gas_limit": "200000"
   }
  },
  "signatures": [
   "0BsM4jdGj/qqgtzMaVZkhDZguX6ol6hL18KxR17yr60MqW9yMNMA8bCwonLEPjSL0PgCzVCll5V9tlfAlKak1g=="
  ]
 },
 "mode": "block"
}
```

## Additional Infomation

### Validator properties

| property          | type     | description                                                                                                       |
| ----------------- | -------- | ----------------------------------------------------------------------------------------------------------------- |
| moniker           | `string` | Identifies your name as seen on the leaderboard table                                                             |
| description       | `string` | Longer description of your node                                                                                   |
| website           | `string` | URL to the validator website                                                                                      |
| avatar            | `string` | URL to image or gif                                                                                               |
| social            | `string` | URL to any social profile such as tweeter or telegram                                                             |
| contact           | `string` | Email address or URL to a submission form                                                                         |
| validator-node-id | `string` | node id of your validator node (required if you want your node to be present in the network visualizer)           |
| sentry-node-id    | `string` | comma separated list of sentry node ids (required if you want your nodes to be present in the network visualizer) |

## Caching Configurations of RPC methods

### `/api/rpc_methods` contains caching information for all endpoints

- caching_blocks: integer defining number of blocks that the response data will  be preserved
- caching_time: integer defining time in seconds that the response data will be preserved for
- caching_enabled: boolean if caching is enabled or not
Here `0` represents no caching, `-1` represents infinit caching, `positive integer` represetns the caching time or number of blocks
Remember this settings when you set/update manually from `config.json`.

### Additionally `config.json` file contains `cache-size-limit` config param which represents the cache size limit in bytes. If cache reaches the limit, it will randomly delete folders/files

### How to update caching configurations

All caching configurations are set in `config.json` file.
`config.json` file includes `rpc_methods` field and there you can set/update caching config of each endpoint.
