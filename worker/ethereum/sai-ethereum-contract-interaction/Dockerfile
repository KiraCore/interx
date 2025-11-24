FROM golang AS build

WORKDIR /src/

COPY ./ /src/

RUN go build -o sai-eth-interaction -buildvcs=false

FROM ubuntu

# Accept VERSION build arg and set as environment variable
ARG VERSION=unknown
ENV VERSION=${VERSION}

WORKDIR /srv

# Copy binary from build stage
COPY --from=build /src/sai-eth-interaction /srv/sai-eth-interaction

# Copy other files
#COPY ./config.yml /srv/config.yml

RUN chmod +x /srv/sai-eth-interaction

# Copy config file
COPY config.yml /srv/config.yml

# Write version to file for easy inspection
RUN echo "${VERSION}" > /srv/VERSION

# Set command to run your binary
CMD /srv/sai-eth-interaction start
