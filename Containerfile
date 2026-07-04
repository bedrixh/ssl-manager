FROM golang:tip-alpine3.24 AS build

WORKDIR /app
RUN apk update
RUN apk add make
COPY . .

RUN make build

FROM alpine:3.23

WORKDIR /app

COPY --from=build /app/bin/ssl-manager /app

CMD ["./ssl-manager", "--daemon", "--config","/app/conf.yaml"]

