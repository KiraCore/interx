#!/usr/bin/env bash
set -e
set -x
. /etc/profile

go mod tidy
go mod verify

PKG_CONFIG_FILE=./nfpm.yaml 
VERSION=$(./scripts/version.sh)

function pcgConfigure() {
    local ARCH="$1"
    local VERSION="$2"
    local PLATFORM="$3"
    local SOURCE="$4"
    local CONFIG="$5"
    SOURCE=${SOURCE//"/"/"\/"}
    sed -i="" "s/\${ARCH}/$ARCH/" $CONFIG
    sed -i="" "s/\${VERSION}/$VERSION/" $CONFIG
    sed -i="" "s/\${PLATFORM}/$PLATFORM/" $CONFIG
    sed -i="" "s/\${SOURCE}/$SOURCE/" $CONFIG
}

function pcgRelease() {
    local ARCH="$1"
    local VERSION="$2"
    local PLATFORM="$3"

    local BIN_PATH=./bin/$ARCH/$PLATFORM
    local RELEASE_DIR=./bin/deb/$PLATFORM
    mkdir -p $BIN_PATH $RELEASE_DIR

    echoInfo "INFO: Building $ARCH package for $PLATFORM..."
    env GOOS=$PLATFORM GOARCH=$ARCH go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o $BIN_PATH
    TMP_PKG_CONFIG_FILE=./nfpm_${ARCH}_${PLATFORM}.yaml
    rm -rfv $TMP_PKG_CONFIG_FILE && cp -v $PKG_CONFIG_FILE $TMP_PKG_CONFIG_FILE

    if [ "${PLATFORM,,}" != "windows" ] ; then
        local RELEASE_PATH="${RELEASE_DIR}/sekai_${VERSION}_${ARCH}.deb"
        pcgConfigure "$ARCH" "$VERSION" "$PLATFORM" "$BIN_PATH" $TMP_PKG_CONFIG_FILE
        nfpm pkg --packager deb --target $RELEASE_PATH -f $TMP_PKG_CONFIG_FILE
        cp -fv "$RELEASE_PATH" ./bin/interx-${PLATFORM}-${ARCH}.deb
    else
        # deb is not supported on windows, simply copy the executables
        cp -fv $BIN_PATH/interx.exe ./bin/interx-${PLATFORM}-${ARCH}.exe
    fi
}

rm -rfv ./bin

# NOTE: To see available build architectures, run: go tool dist list
pcgRelease "amd64" "$VERSION" "linux"
pcgRelease "amd64" "$VERSION" "darwin"
pcgRelease "arm64" "$VERSION" "linux"
pcgRelease "arm64" "$VERSION" "darwin"

rm -rfv ./bin/amd64 ./bin/arm64 ./bin/deb
echoInfo "INFO: Sucessfully published INTERX deb packages into ./bin"
