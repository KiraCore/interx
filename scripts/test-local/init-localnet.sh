#!/usr/bin/env bash
# To run test locally: make network-start && ./scripts/test-local/token-transfers.sh
set -e
set -x
. /etc/profile

INTERX_VERSION=$(./scripts/version.sh InterxVersion)
SEKAID_VERSION=$(./scripts/version.sh SekaiVersion)
DEFAULT_GRPC_PORT=9090
DEFAULT_RPC_PORT=26657
DEFAULT_INTERX_PORT=11000
PING_TARGET="127.0.0.1"
CFG_grpc="dns:///$PING_TARGET:$DEFAULT_GRPC_PORT"
CFG_rpc="http://$PING_TARGET:$DEFAULT_RPC_PORT"
INTERXD_HOME=$HOME/.interxd

interxd init --cache_dir="$INTERXD_HOME/cache" --home="$INTERXD_HOME" --grpc="$CFG_grpc" --rpc="$CFG_rpc" --port="$INTERNAL_API_PORT" \
    --signing_mnemonic="$INTERXD_HOME/interx.mnemonic" \
    --faucet_mnemonic="$INTERXD_HOME/faucet.mnemonic" \
    --port="$DEFAULT_INTERX_PORT" \
    --node_type="validator" \
    --seed_node_id="" \
    --sentry_node_id="" \
    --validator_node_id="$(globGet validator_node_id)" \
    --addrbook="$(globFile KIRA_ADDRBOOK)" \
    --faucet_time_limit=30 \
    --faucet_amounts="100000ukex,20000000test,300000000000000000samolean,1lol" \
    --faucet_minimum_amounts="1000ukex,50000test,250000000000000samolean,1lol" \
    --fee_amounts="ukex 1000ukex,test 500ukex,samolean 250ukex,lol 100ukex"
