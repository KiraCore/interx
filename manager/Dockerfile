FROM golang AS build

WORKDIR /src

COPY ./ /src/

RUN go build -o sai-interx-manager -buildvcs=false

FROM ubuntu

WORKDIR /srv
RUN apt-get update && apt-get -y install ca-certificates

COPY --from=build /src/sai-interx-manager /srv/

RUN chmod +x /srv/sai-interx-manager

CMD /srv/sai-interx-manager start > /srv/logs/app.log 2>&1
