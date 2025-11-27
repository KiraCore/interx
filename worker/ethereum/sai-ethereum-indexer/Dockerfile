# Build
FROM golang AS build

WORKDIR /src/cmd/app

COPY ./ /src/

RUN go build -o sai-eth-indexer -buildvcs=false

FROM ubuntu

# Accept VERSION build arg and set as environment variable
ARG VERSION=unknown
ENV VERSION=${VERSION}

WORKDIR /srv
RUN apt-get update && apt-get -y install ca-certificates

# Copy binary from build stage
COPY --from=build /src/cmd/app/sai-eth-indexer /srv/

RUN chmod +x /srv/sai-eth-indexer

# Copy config file
COPY config.json /srv/config.json

# Write version to file for easy inspection
RUN echo "${VERSION}" > /srv/VERSION

# Set command to run your binary
CMD /srv/sai-eth-indexer --debug
