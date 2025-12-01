# Build stage
FROM golang AS build

WORKDIR /src/

COPY ./ /src/

ENV GOPRIVATE=github.com/KiraCore/*
RUN go mod tidy && go build -o sai-storage-bin -buildvcs=false

FROM ubuntu

# Accept VERSION build arg and set as environment variable
ARG VERSION=unknown
ENV VERSION=${VERSION}

WORKDIR /srv

# Copy binary from build stage
COPY --from=build /src/sai-storage-bin /srv/sai-storage-bin

# Copy other files
#COPY ./config.yml /srv/config.yml

RUN chmod +x /srv/sai-storage-bin

# Copy config file
COPY config.yml /srv/config.yml

# Write version to file for easy inspection
RUN echo "${VERSION}" > /srv/VERSION

# Set command to run your binary
CMD /srv/sai-storage-bin start

EXPOSE 8880
