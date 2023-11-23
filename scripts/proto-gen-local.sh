#!/usr/bin/env bash
set -e
set -x
. /etc/profile

proto_dirs=$(find ./proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
echoInfo "Generating protobuf files..."
for dir in $proto_dirs; do
    proto_fils=$(find "${dir}" -maxdepth 1 -name '*.proto') 
    for fil in $proto_fils; do
        if grep -q "option go_package" "$fil"; then
            buf protoc \
            -I "./proto" \
            -I third_party/grpc-gateway/ \
            -I third_party/googleapis/ \
            -I third_party/proto/ \
            --go_out=paths=source_relative:./proto-gen \
            --go-grpc_out=paths=source_relative:./proto-gen \
            --grpc-gateway_out=logtostderr=true,paths=source_relative:./proto-gen \
            $fil || ( echoErr "ERROR: Failed proto build for: ${fil}" && sleep 2 && exit 1 )
        fi
    done
done

# protogen_dirs=$(find ./proto-gen -path -prune -o -name '*.gw.go' -print0 | xargs -0 -n1 dirname | sort | uniq)

# echoInfo "Updating proto generated files to include relative paths..."
# for dir in $protogen_dirs; do
#     protogen_fils=$(find "${dir}" -maxdepth 1 -name '*.gw.go') 
#     for fil in $protogen_fils; do
#         sed -i="" 's/github.com\/grpc-ecosystem\/grpc-gateway\/runtime/github.com\/grpc-ecosystem\/grpc-gateway\/v2\/runtime/g' "$fil" || ( echoErr "ERROR: Failed to sed file: '$fil'" && exit 1 )
#         sed -i="" 's/github.com\/grpc-ecosystem\/grpc-gateway\/utilities/github.com\/grpc-ecosystem\/grpc-gateway\/v2\/utilities/g' "$fil" || ( echoErr "ERROR: Failed to sed file: '$fil'" && exit 1 )
#     done
# done

#TODO: Currently it is not possible for go to dicover the gocosmos_out plugin (might require to resolve some issues with path)
#--gocosmos_out=plugins=interfacetype+grpc,\
#Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:./proto-gen \

echoInfo "INFO: Success, all proto files were compiled!"
