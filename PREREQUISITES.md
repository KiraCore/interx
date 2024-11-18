## Prerequisites

```
# install essential dependencies
apt-get install -y curl && curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add - && apt-get update -y && \
 apt-get install -y --allow-unauthenticated --allow-downgrades --allow-remove-essential --allow-change-held-packages \
 software-properties-common wget git nginx apt-transport-https file build-essential net-tools hashdeep \
 protobuf-compiler golang-goprotobuf-dev golang-grpc-gateway golang-github-grpc-ecosystem-grpc-gateway-dev lsb-release \
 clang cmake gcc g++ pkg-config libudev-dev libusb-1.0-0-dev iputils-ping nano jq python python3 python3-pip gnupg \
 bash libglu1-mesa lsof bc dnsutils psmisc netcat  make nodejs tar unzip xz-utils yarn zip p7zip-full ca-certificates \
 containerd docker.io dos2unix

# install console helper
KIRA_TOOLS_BRANCH="v0.1.0.7" && cd /tmp && rm -fv ./i.sh && \
    wget https://github.com/KiraCore/tools/releases/download/$KIRA_TOOLS_BRANCH/bash-utils.sh -O ./i.sh && \
    chmod 777 ./i.sh && ./i.sh bashUtilsSetup "/var/kiraglob" && . /etc/profile

# install deb package manager
echo 'deb [trusted=yes] https://repo.goreleaser.com/apt/ /' | tee /etc/apt/sources.list.d/goreleaser.list && apt-get update -y && \
	apt install nfpm
```