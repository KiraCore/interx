FROM golang AS build

WORKDIR /src

COPY ./ /src/

RUN go build -o sai-interax-proxy -buildvcs=false

FROM ubuntu

# Accept VERSION build arg and set as environment variable
ARG VERSION=unknown
ENV VERSION=${VERSION}

WORKDIR /srv
RUN apt-get update && apt-get -y install ca-certificates

COPY --from=build /src/sai-interax-proxy /srv/
COPY --from=build /src/config.yml /srv/config.yml

RUN chmod +x /srv/sai-interax-proxy

RUN echo "${VERSION}" > /srv/VERSION
CMD /srv/sai-interax-proxy start
