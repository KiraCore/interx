## Signatures :pen:

Cosign requires simple initial setup of the signer keys described more precisely [here](https://dev.to/n3wt0n/sign-your-container-images-with-cosign-github-actions-and-github-container-registry-3mni). You can use the one-liner command below to get up to speed fast.

```bash
# install cosign
COSIGN_VERSION="v1.7.2" && \
if [[ "$(uname -m)" == *"ar"* ]] ; then ARCH="arm64"; else ARCH="amd64" ; fi && echo $ARCH && \
PLATFORM=$(uname) && FILE=$(echo "cosign-${PLATFORM}-${ARCH}" | tr '[:upper:]' '[:lower:]') && \
 wget https://github.com/sigstore/cosign/releases/download/${COSIGN_VERSION}/$FILE && chmod +x -v ./$FILE && \
 mv -fv ./$FILE /usr/local/bin/cosign && cosign version

# save KIRA public cosign key
KEYS_DIR="/usr/keys" && KIRA_COSIGN_PUB="${KEYS_DIR}/kira-cosign.pub" && \
mkdir -p $KEYS_DIR  && cat > ./cosign.pub << EOL
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE/IrzBQYeMwvKa44/DF/HB7XDpnE+
f+mU9F/Qbfq25bBWV2+NlYMJv3KvKHNtu3Jknt6yizZjUV4b8WGfKBzFYw==
-----END PUBLIC KEY-----
EOL

# download desired files and the corresponding .sig file from: https://github.com/KiraCore/interx/releases

# verify signature of downloaded files
cosign verify-blob --key=$KIRA_COSIGN_PUB--signature=./<file>.sig ./<file>
```