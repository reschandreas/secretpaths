FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN rm -rf demo/

ENV CGO_ENABLED=0 \
    GOOS=linux

RUN go build -o /server

FROM scratch
ENV GIN_MODE=release
ENV VAULT_ADDR=http://localhost:8200
ENV VAULT_KV_ENGINE_PATH=secret
COPY --from=builder /server /server

EXPOSE 8081

# Run
CMD ["/server"]