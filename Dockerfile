# syntax=docker/dockerfile:1

ARG BITCOIN_IMAGE=ruimarinho/bitcoin-core:27

FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG APP
RUN test -n "$APP"
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/app ./cmd/${APP}

FROM ${BITCOIN_IMAGE}

USER root
WORKDIR /data

COPY --from=build /out/app /usr/local/bin/app

EXPOSE 8080 8081 8082

ENTRYPOINT ["/usr/local/bin/app"]
