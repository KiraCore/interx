#!/bin/bash
set -e
set -x
. /etc/profile

CURRENT_DIR=$(pwd)
UTILS_VER=$(utilsVersion 2> /dev/null || echo "")
GO_VER=$(go version 2> /dev/null || echo "")

# Installing utils is essential to simplify the setup steps
if [ -z "$UTILS_VER" ] ; then
    echo "INFO: KIRA utils were NOT installed on the system, setting up..."
    KIRA_UTILS_BRANCH="v0.0.2" && cd /tmp && rm -fv ./i.sh && \
    wget https://raw.githubusercontent.com/KiraCore/tools/$KIRA_UTILS_BRANCH/bash-utils/install.sh -O ./i.sh && \
    chmod 555 -v ./i.sh && ./i.sh "$KIRA_UTILS_BRANCH" "/var/kiraglob" && loadGlobEnvs
fi

# install golang if needed
if  ($(isNullOrEmpty "$GO_VER")) ; then
    GO_VERSION="1.17.7" && ARCH=$(([[ "$(uname -m)" == *"arm"* ]] || [[ "$(uname -m)" == *"aarch"* ]]) && echo "arm64" || echo "amd64") && \
     GO_TAR=go${GO_VERSION}.linux-${ARCH}.tar.gz && rm -rfv /usr/local/go && cd /tmp && rm -fv ./$GO_TAR && \
     wget https://dl.google.com/go/${GO_TAR} && \
     tar -C /usr/local -xvf $GO_TAR && \
     setGlobEnv GOROOT "/usr/local/go" && setGlobPath "\$GOROOT" && \
     setGlobEnv GOBIN "/usr/local/go/bin" && setGlobPath "\$GOBIN" && \
     setGlobEnv GOPATH "/home/go" && setGlobPath "\$GOPATH" && \
     setGlobEnv GOCACHE "/home/go/cache" && \
     loadGlobEnvs && \
     mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

    echoInfo "INFO: Sucessfully intalled $(go version)"
fi

# navigate to current direcotry and load global environment variables
cd $CURRENT_DIR
loadGlobEnvs

if ($(isNullOrEmpty "$SEKAI_BRANCH")) ; then
    SEKAI_BRANCH="master"
    echoWarn "WARNING: SEKAI branch 'SEKAI_BRANCH' env variable was undefined, the '$SEKAI_BRANCH' branch will be used during installation process!" && sleep 1
    setGlobEnv SEKAI_BRANCH "$SEKAI_BRANCH"
fi

if ($(isNullOrEmpty "$GOBIN")) ; then
    GOBIN=${HOME}/go/bin
    echoWarn "WARNING: GOBIN env variable was undefined, the '$GOBIN' will be used during installation process!" && sleep 1
fi

go clean -modcache
BUF_VER=$(buf --version 2> /dev/null || echo "")

if ($(isNullOrEmpty "$BUF_VER")) ; then
    GO111MODULE=on 
    go install github.com/bufbuild/buf/cmd/buf@v1.0.0-rc10
    echoInfo "INFO: Sucessfully intalled buf $(buf --version)"

    setGlobEnv GOLANG_PROTOBUF_VERSION "1.27.1" && \
     setGlobEnv GOGO_PROTOBUF_VERSION "1.3.2" && \
     setGlobEnv GRPC_GATEWAY_VERSION "1.14.7" && \
     loadGlobEnvs

    go install github.com/cosmos/cosmos-proto/cmd/protoc-gen-go-pulsar@latest && \
     go install google.golang.org/protobuf/cmd/protoc-gen-go@v${GOLANG_PROTOBUF_VERSION} && \
     go install github.com/gogo/protobuf/protoc-gen-gogo@v${GOGO_PROTOBUF_VERSION} && \
     go install github.com/gogo/protobuf/protoc-gen-gogofast@v${GOGO_PROTOBUF_VERSION} && \
     go install github.com/gogo/protobuf/protoc-gen-gogofaster@v${GOGO_PROTOBUF_VERSION} && \
     go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v${GRPC_GATEWAY_VERSION} && \
     go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v${GRPC_GATEWAY_VERSION}

    # Following command executes with error requiring us to silence it, however the executable is placed in $GOBIN
    # https://github.com/regen-network/cosmos-proto
    # reference: 
    go install github.com/regen-network/cosmos-proto/protoc-gen-gocosmos@v0.3.1 2> /dev/null || : 
    go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest

    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
fi

function goUpdate(){
    local module=$1
    local version=$2
    local current=$(go list -m -f '{{ .Version }}' github.com/KiraCore/sekai 2> /dev/null || echo "")
    if [ "$current" != "$version" ] ; then
        go get "${module}@${version}"
        echoInfo "INFO: Updated module ${module}, $current -> $version"
    else
        echoInfo "INFO: Module ${module}@${version} is up to date"
    fi
}

COSMOS_BRANCH="v0.45.1"
SEKAI_BRANCH="master"
goUpdate github.com/KiraCore/sekai "$SEKAI_BRANCH"
goUpdate github.com/cosmos/cosmos-sdk "$COSMOS_BRANCH"

echoInfo "Cleaning up proto gen files..."
rm -rfv ./proto-gen
mkdir -p ./proto-gen ./proto
kira_dir=$(go list -f '{{ .Dir }}' -m github.com/KiraCore/sekai@$SEKAI_BRANCH)
cosmos_sdk_dir=$(go list -f '{{ .Dir }}' -m github.com/cosmos/cosmos-sdk@$COSMOS_BRANCH)

rm -rfv ./proto/cosmos ./proto/kira ./third_party/proto
mkdir -p ./third_party/proto
cp -rfv $cosmos_sdk_dir/proto/cosmos ./proto
cp -rfv $cosmos_sdk_dir/third_party/proto/cosmos_proto ./third_party/proto
cp -rfv $cosmos_sdk_dir/third_party/proto/tendermint ./third_party/proto
cp -rfv $kira_dir/proto/kira ./proto

proto_dirs=$(find ./proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)

echoInfo "Generating protobuf files..."

for dir in $proto_dirs; do
    proto_fils=$(find "${dir}" -maxdepth 1 -name '*.proto') 
    for fil in $proto_fils; do
        buf protoc \
          -I "./proto" \
          -I third_party/grpc-gateway/ \
		  -I third_party/googleapis/ \
		  -I third_party/proto/ \
          --go_out=paths=source_relative:./proto-gen \
          --go-grpc_out=paths=source_relative:./proto-gen \
          --grpc-gateway_out=paths=source_relative:./proto-gen \
          $fil || ( echoErr "ERROR: Failed proto build for: ${fil}" && sleep 1 )
    done
done

echoInfo "Proto files were generated for:"
echoInfo echo ${proto_dirs[*]}
sleep 1

go mod tidy
go build -o "${GOBIN}/interxd"
go mod verify
echoInfo "INFO: Sucessfully intalled INTERX $(interxd version)"
