FROM golang AS build

WORKDIR /src

COPY ./ /src/

ENV GOPRIVATE=github.com/KiraCore/*
RUN go mod tidy && go build -o sai-interx-manager -buildvcs=false

FROM ubuntu

# Accept VERSION build arg and set as environment variable
ARG VERSION=unknown
ENV VERSION=${VERSION}

WORKDIR /srv
RUN apt-get update && apt-get -y install ca-certificates

COPY --from=build /src/sai-interx-manager /srv/
COPY --from=build /src/config.yml /srv/config.yml

RUN chmod +x /srv/sai-interx-manager

RUN echo "${VERSION}" > /srv/VERSION

CMD /srv/sai-interx-manager start

