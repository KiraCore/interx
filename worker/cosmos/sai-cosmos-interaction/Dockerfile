# Build
FROM golang AS build

WORKDIR /src

COPY ./ /src/

ENV GOPRIVATE=github.com/KiraCore/*
RUN go mod tidy && go build -o sai-cosmos-interaction -buildvcs=false

FROM ubuntu

# Accept VERSION build arg and set as environment variable
ARG VERSION=unknown
ENV VERSION=${VERSION}

WORKDIR /srv
RUN apt-get update && apt-get -y install ca-certificates

# Copy binary from build stage
COPY --from=build /src/sai-cosmos-interaction /srv/

RUN chmod +x /srv/sai-cosmos-interaction

# Copy config file
COPY config.yml /srv/config.yml

# Write version to file for easy inspection
RUN echo "${VERSION}" > /srv/VERSION

# Set command to run your binary
CMD /srv/sai-cosmos-interaction start
