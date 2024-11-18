# Query Block by Number (note that you do NOT need to query status to get the next block, because its hash is included in the previous getblock querry)
RPC_URL="127.0.0.1:18332" && RPC_CRED="admin:1234"
# Test Legacy Address: 2NFM1PTfd49rWfNK2xc5CiK3kYuReKfm49K, 2MubTWazTjHy2ZfRAyUo3giMJiimKojNmaJ
# Test Segwit Address: tb1qremdntfgz7nfdk8fpx8np5f9hzq3vjam0rwc8f
# Test Multisig Address: 2N1CqAMk642sb77PdUHufDDQCw6qNEDnhrS

# validate correctness of an adresses that should be added to monitoring
ADDR="2N1CqAMk642sb77PdUHufDDQCw6qNEDnhrS" && \
curl --user "$RPC_CRED" --data-binary "{\"jsonrpc\": \"2.0\", \"method\": \"validateaddress\", \"params\": [ \"$ADDR\" ]}" "$RPC_URL" | jq ".result"

# # Wallet RPC calls wil NOT work if a temporary 'test' wallet is not created first, the test wallet must have both 'disable_private_keys' and 'blank' arguments set to 'true'
# # Note that by default all new wallets are descriptor wallets and that does NOT allow to use importaddress 'importaddress', for importing legacy & segwith all params must be supplied
# # See createwallet reference: https://bitcoincore.org/en/doc/22.0.0/rpc/wallet/createwallet/
# # set load on startup argument #7 to true
# ADDR="tb1qzg0c8u8hxlcgnvj0sx5g7tq908phkldxv2tmyx" && \
# curl --user "$RPC_CRED" --data-binary "{\"jsonrpc\": \"2.0\", \"method\": \"createwallet\", \"params\": [ \"$ADDR\", true, true, \"\", false, false, true ]}" "$RPC_URL" | jq

# To check if wallet was already created or not use listwallets
curl --user "$RPC_CRED" --data-binary "{\"jsonrpc\": \"2.0\", \"method\": \"listwallets\", \"params\": [ ]}" "$RPC_URL" | jq ".result"

# # See importaddress reference: https://bitcoincore.org/en/doc/22.0.0/rpc/wallet/importaddress/
# # The p2sh argument should only be used with scripts (e.g. multisig)
# # The rescan argument should be set to false otherwise the RPC querry will be pending until scan finishes running
# ADDR="tb1qzg0c8u8hxlcgnvj0sx5g7tq908phkldxv2tmyx" && \
# curl --user "$RPC_CRED" --data-binary "{\"jsonrpc\": \"2.0\", \"method\": \"importaddress\", \"params\": [ \"$ADDR\", \"\", false ]}" "$RPC_URL/wallet/$ADDR" | jq

# # Ensure that wallet file is loaded
# ADDR="tb1qzg0c8u8hxlcgnvj0sx5g7tq908phkldxv2tmyx" && \
# curl --user "$RPC_CRED" --data-binary "{\"jsonrpc\": \"2.0\", \"method\": \"loadwallet\", \"params\": [ \"$ADDR\", true ]}" "$RPC_URL" | jq

# # Get address info from the wallet
# ADDR="2N1CqAMk642sb77PdUHufDDQCw6qNEDnhrS" && \
# curl --user "$RPC_CRED" --data-binary "{\"jsonrpc\": \"2.0\", \"method\": \"getaddressinfo\", \"params\": [ \"$ADDR\" ]}" "$RPC_URL/wallet/$ADDR" | jq ".result"

# # Get wallet info
# ADDR="2N1CqAMk642sb77PdUHufDDQCw6qNEDnhrS" && \
# curl --user "$RPC_CRED" --data-binary "{\"jsonrpc\": \"2.0\", \"method\": \"getwalletinfo\", \"params\": [ ]}" "$RPC_URL/wallet/$ADDR" | jq ".result"

# # Rescan of a specific wallet is essential to keep track of balances and transactions
# ADDR="tb1qzg0c8u8hxlcgnvj0sx5g7tq908phkldxv2tmyx" && \
# curl --user "$RPC_CRED" --data-binary "{\"jsonrpc\": \"2.0\", \"method\": \"rescanblockchain\", \"params\": [ 0 ]}" "$RPC_URL/wallet/$ADDR" | jq ".result"

# # Get wallet balances (this requires re-scan)
# ADDR="tb1qzg0c8u8hxlcgnvj0sx5g7tq908phkldxv2tmyx" && \
# curl --user "$RPC_CRED" --data-binary "{\"jsonrpc\": \"2.0\", \"method\": \"getbalances\", \"params\": [ ]}" "$RPC_URL/wallet/$ADDR" | jq ".result"

# # Get wallet transactions
# ADDR="tb1qzg0c8u8hxlcgnvj0sx5g7tq908phkldxv2tmyx" && \
# curl --user "$RPC_CRED" --data-binary "{\"jsonrpc\": \"2.0\", \"method\": \"listtransactions\", \"params\": [ \"*\", 2147483647, 0, true ]}" "$RPC_URL/wallet/$ADDR" | jq

# # Get unspent transactions, ref.: https://developer.bitcoin.org/reference/rpc/listunspent.html
# ADDR="tb1qzg0c8u8hxlcgnvj0sx5g7tq908phkldxv2tmyx" && \
# curl --user "$RPC_CRED" --data-binary "{\"jsonrpc\": \"2.0\", \"method\": \"listunspent\", \"params\": [ 1, 2147483647, [ \"$ADDR\" ], true ]}" "$RPC_URL/wallet/$ADDR" | jq

# # Get list of addresses in the wallet (Note that this will ONLY return entries IF AND ONLY IF rescan was finalized and addresses have non 0 balances)
# ADDR="tb1qzg0c8u8hxlcgnvj0sx5g7tq908phkldxv2tmyx" && \
# curl --user "$RPC_CRED" --data-binary "{\"jsonrpc\": \"2.0\", \"method\": \"listreceivedbyaddress\", \"params\": [ 0, true, true ]}" "$RPC_URL/wallet/$ADDR" | jq ".result"

ADDR="tb1qzg0c8u8hxlcgnvj0sx5g7tq908phkldxv2tmyx" && \
curl --user "$RPC_CRED" --data-binary "{\"jsonrpc\": \"2.0\", \"method\": \"listaddressgroupings\", \"params\": []}" "$RPC_URL/wallet/$ADDR" | jq ".result"
