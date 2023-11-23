#!/usr/bin/env bash
set -e
set -x
. /etc/profile

CURRENT_DIR=$(pwd)
GO_VER=$(go version 2> /dev/null || echo "")

UTILS_VER=$(bashUtilsVersion 2> /dev/null || echo "")
UTILS_OLD_VER="false" && [[ $(versionToNumber "$UTILS_VER" || echo "0") -ge $(versionToNumber "v0.1.2.3" || echo "1") ]] || UTILS_OLD_VER="true" 

# Installing utils is essential to simplify the setup steps
if [ "$UTILS_OLD_VER" == "true" ] ; then
    echo "INFO: KIRA utils were NOT installed on the system, setting up..." && sleep 2
    KIRA_TOOLS_BRANCH="v0.1.5" && cd /tmp && rm -fv ./i.sh && \
    wget https://github.com/KiraCore/tools/releases/download/$KIRA_TOOLS_BRANCH/bash-utils.sh -O ./i.sh && \
    chmod 777 ./i.sh && ./i.sh bashUtilsSetup "/var/kiraglob" && . /etc/profile
else
    echoInfo "INFO: KIRA utils are up to date, latest version $UTILS_VER" && sleep 2
fi

# install golang if needed
if  ($(isNullOrEmpty "$GO_VER")) || ($(isNullOrEmpty "$GOBIN")) ; then
    GO_VERSION="1.19" && ARCH=$(([[ "$(uname -m)" == *"arm"* ]] || [[ "$(uname -m)" == *"aarch"* ]]) && echo "arm64" || echo "amd64") && \
     OS_VERSION=$(uname) && GO_TAR=go${GO_VERSION}.${OS_VERSION,,}-${ARCH}.tar.gz && rm -rfv /usr/local/go && cd /tmp && rm -fv ./$GO_TAR && \
     wget https://dl.google.com/go/${GO_TAR} && \
     tar -C /usr/local -xvf $GO_TAR && rm -fv ./$GO_TAR && \
     setGlobEnv GOROOT "/usr/local/go" && setGlobPath "\$GOROOT" && \
     setGlobEnv GOBIN "/usr/local/go/bin" && setGlobPath "\$GOBIN" && \
     setGlobEnv GOPATH "$HOME/go" && setGlobPath "\$GOPATH" && \
     setGlobEnv GOCACHE "$HOME/go/cache" && \
     loadGlobEnvs && \
     mkdir -p "$GOPATH/src" "$GOPATH/bin" "$GOCACHE" && \
     chmod -R 777 "$GOPATH" && chmod -R 777 "$GOROOT" && chmod -R 777 "$GOCACHE"

    echoInfo "INFO: Sucessfully intalled $(go version)"
fi

# navigate to current direcotry and load global environment variables
cd $CURRENT_DIR
loadGlobEnvs

go clean -modcache
EXPECTED_INTERX_PROTO_DEP_VER="v0.0.6"
BUF_VER=$(buf --version 2> /dev/null || echo "")

if ($(isNullOrEmpty "$BUF_VER")) || [ "$INTERX_PROTO_DEP_VER" != "$EXPECTED_INTERX_PROTO_DEP_VER" ] ; then

    rm -f /usr/local/go/bin/buf
    rm -f /usr/local/bin/buf
    rm -f /usr/bin/buf
    rm -f $HOME/go/bin/buf
    rm -f $GOBIN/buf
    # which buf

    rm -f /usr/local/go/bin/protoc-gen-go-grpc
    rm -f /usr/local/bin/protoc-gen-go-grpc
    rm -f /usr/bin/protoc-gen-go-grpc
    rm -f $HOME/go/bin/protoc-gen-go-grpc
    rm -f $GOBIN/protoc-gen-go-grpc
    # which protoc-gen-go-grpc

    rm -f /usr/local/go/bin/protoc-gen-doc
    rm -f /usr/local/bin/protoc-gen-doc
    rm -f /usr/bin/protoc-gen-doc
    rm -f $HOME/go/bin/protoc-gen-doc
    rm -f $GOBIN/protoc-gen-doc
    # which protoc-gen-doc

    rm -f /usr/local/go/bin/protoc-gen-gogotypes
    rm -f /usr/local/bin/protoc-gen-gogotypes
    rm -f /usr/bin/protoc-gen-gogotypes
    rm -f $HOME/go/bin/protoc-gen-gogotypes
    rm -f $GOBIN/protoc-gen-gogotypes
    # which protoc-gen-gogotypes

    rm -f /usr/local/go/bin/protoc-gen-grpc-gateway
    rm -f /usr/local/bin/protoc-gen-grpc-gateway
    rm -f /usr/bin/protoc-gen-grpc-gateway
    rm -f $HOME/go/bin/protoc-gen-grpc-gateway
    rm -f $GOBIN/protoc-gen-grpc-gateway
    # which protoc-gen-grpc-gateway

    rm -f /usr/local/go/bin/protoc-gen-gogofaster
    rm -f /usr/local/bin/protoc-gen-gogofaster
    rm -f /usr/bin/protoc-gen-gogofaster
    rm -f $HOME/go/bin/protoc-gen-gogofaster
    rm -f $GOBIN/protoc-gen-gogofaster
    # which protoc-gen-gogofaster

    rm -f /usr/local/go/bin/protoc-gen-gogofast
    rm -f /usr/local/bin/protoc-gen-gogofast
    rm -f /usr/bin/protoc-gen-gogofast
    rm -f $HOME/go/bin/protoc-gen-gogofast
    rm -f $GOBIN/protoc-gen-gogofast
    # which protoc-gen-gogofast

    rm -f /usr/local/go/bin/protoc-gen-gogo
    rm -f /usr/local/bin/protoc-gen-gogo
    rm -f /usr/bin/protoc-gen-gogo
    rm -f $HOME/go/bin/protoc-gen-gogo
    rm -f $GOBIN/protoc-gen-gogo
    # which protoc-gen-gogo

    rm -f /usr/local/go/bin/protoc-gen-go
    rm -f /usr/local/bin/protoc-gen-go
    rm -f /usr/bin/protoc-gen-go
    rm -f $HOME/go/bin/protoc-gen-go
    rm -f $GOBIN/protoc-gen-go
    # which protoc-gen-go

    rm -f /usr/local/go/bin/protoc-gen-go-pulsar
    rm -f /usr/local/bin/protoc-gen-go-pulsar
    rm -f /usr/bin/protoc-gen-go-pulsar
    rm -f $HOME/go/bin/protoc-gen-go-pulsar
    rm -f $GOBIN/protoc-gen-go-pulsar
    # which protoc-gen-go-pulsar

    GO111MODULE=on 
    go install github.com/bufbuild/buf/cmd/buf@v1.0.0-rc10
    echoInfo "INFO: Sucessfully intalled buf $(buf --version)"

    setGlobEnv GOLANG_PROTOBUF_VERSION "1.28.0" && \
     setGlobEnv GOGO_PROTOBUF_VERSION "1.3.2" && \
     setGlobEnv GRPC_GATEWAY_VERSION "2.10.3" && \
     loadGlobEnvs

    go install github.com/cosmos/cosmos-proto/cmd/protoc-gen-go-pulsar@latest && \
     go install google.golang.org/protobuf/cmd/protoc-gen-go@v${GOLANG_PROTOBUF_VERSION} && \
     go install github.com/gogo/protobuf/protoc-gen-gogo@v${GOGO_PROTOBUF_VERSION} && \
     go install github.com/gogo/protobuf/protoc-gen-gogofast@v${GOGO_PROTOBUF_VERSION} && \
     go install github.com/gogo/protobuf/protoc-gen-gogofaster@v${GOGO_PROTOBUF_VERSION} && \
     go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v${GRPC_GATEWAY_VERSION} && \
    #  go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-swagger@v${GRPC_GATEWAY_VERSION} && \
     go install github.com/gogo/protobuf/protoc-gen-gogotypes@v1.3.2

    # Following command executes with error requiring us to silence it, however the executable is placed in $GOBIN
    # https://github.com/regen-network/cosmos-proto
    # Original setup originates from Docker Image tendermintdev/sdk-proto-gen:v0.2
    # reference: 
    go install github.com/regen-network/cosmos-proto/protoc-gen-gocosmos@v0.3.1 2> /dev/null || : 
    go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest

    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
    setGlobEnv INTERX_PROTO_DEP_VER "$EXPECTED_INTERX_PROTO_DEP_VER"
fi

CONSTANS_FILE=./config/constants.go
COSMOS_BRANCH=$(grep -Fn -m 1 'CosmosVersion ' $CONSTANS_FILE | rev | cut -d "=" -f1 | rev | xargs | tr -dc '[:alnum:]\-\.' || echo '')
SEKAI_BRANCH=$(grep -Fn -m 1 'SekaiVersion ' $CONSTANS_FILE | rev | cut -d "=" -f1 | rev | xargs | tr -dc '[:alnum:]\-\.' || echo '')
($(isNullOrEmpty "$COSMOS_BRANCH")) && ( echoErr "ERROR: CosmosVersion was NOT found in contants '$CONSTANS_FILE' !" && sleep 5 && exit 1 )
($(isNullOrEmpty "$SEKAI_BRANCH")) && ( echoErr "ERROR: SekaiVersion was NOT found in contants '$CONSTANS_FILE' !" && sleep 5 && exit 1 )

go get github.com/cosmos/cosmos-sdk@$COSMOS_BRANCH

echoInfo "Cleaning up proto gen files..."
rm -rfv ./proto-gen
mkdir -p ./proto-gen
mkdir -p ./proto-gen ./proto
cosmos_sdk_dir=$(go list -f '{{ .Dir }}' -m github.com/cosmos/cosmos-sdk@$COSMOS_BRANCH)

rm -rfv ./third_party/proto ./proto/cosmos ./proto/kira
mkdir -p ./third_party/proto
cp -rfv $cosmos_sdk_dir/proto/cosmos ./proto
cp -rfv $cosmos_sdk_dir/proto/amino ./proto
tar -C ./third_party/ -xvf ./third_party_cosmos_proto.tar.xz

wget https://github.com/KiraCore/sekai/releases/download/$SEKAI_BRANCH/source-code.tar.gz
mkdir -p sekai
tar -C ./sekai/ -xvf ./source-code.tar.gz
rm -rfv ./source-code.tar.gz
cp -rfv ./sekai/proto/kira ./proto
rm -rfv ./sekai

### This part is required by gocosmos_out
rm -rfv ./codec && mkdir -p codec/types
buf protoc -I "third_party/proto" --gogotypes_out=./codec/types third_party/proto/google/protobuf/any.proto
mv codec/types/google.golang.org/protobuf/types/known/anypb/any.pb.go codec/types
rm -rfv codec/types/google.golang.org/protobuf/types/known/anypb
rm -rfv ./third_party/proto/gogoproto
rm -rfv ./third_party/proto/google
###

sed '/proto\.RegisterType/d' codec/types/any.pb.go > tmp && mv tmp codec/types/any.pb.go

proto_dirs=$(find ./proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)

echoInfo "Updating proto files to include relative paths..."
fil=./proto/cosmos/base/v1beta1/coin.proto && \
 sed -i="" 's/ = \"github.com\/cosmos\/cosmos-sdk\/types/ = \"github.com\/KiraCore\/interx\/proto-gen\/cosmos\/base\/v1beta1/g' "$fil" || ( echoErr "ERROR: Failed to sed file: '$fil'" && exit 1 )
for dir in $proto_dirs; do
    proto_fils=$(find "${dir}" -maxdepth 1 -name '*.proto') 
    for fil in $proto_fils; do
        sed -i="" 's/, (gogoproto.castrepeated) = \"github.com\/cosmos\/cosmos-sdk\/types.Coins\"//g' "$fil" || ( echoErr "ERROR: Failed to sed file: '$fil'" && exit 1 )
        sed -i="" 's/github.com\/cosmos\/cosmos-sdk\/x/github.com\/KiraCore\/interx\/proto-gen\/cosmos/g' "$fil" || ( echoErr "ERROR: Failed to sed file: '$fil'" && exit 1 )
        sed -i="" 's/\[(gogoproto.stdtime) = true, (gogoproto.nullable) = false\]/\[(gogoproto.stdtime) = true, (gogoproto.nullable) = false, (gogoproto.moretags) = \"yaml:\\\"date\\\"\"\]/g' "$fil" || ( echoErr "ERROR: Failed to sed file: '$fil'" && exit 1 )
        sed -i="" 's/github.com\/KiraCore\/interx\/proto-gen\/cosmos\/auth\/types/github.com\/KiraCore\/interx\/proto-gen\/cosmos\/auth\/v1beta1/g' "$fil" || ( echoErr "ERROR: Failed to sed file: '$fil'" && exit 1 )
    done
done

sed -i="" 's/message IdentityRecord {/message IdentityRecord \{\n  option (gogoproto.goproto_getters) = false;/g' ./proto/kira/gov/identity_registrar.proto || ( echoErr "ERROR: Failed to sed file: '$fil'" && exit 1 )
sed -i="" 's/ \[(cosmos_proto.accepts_interface) = \"AccountI\"\]//g' ./proto/cosmos/auth/v1beta1/query.proto || ( echoErr "ERROR: Failed to sed file: '$fil'" && exit 1 )

### This part is required to override some cosmos protos
cp -rfv ./proto-override/cosmos/auth/v1beta1/auth.proto ./proto/cosmos/auth/v1beta1
###

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
