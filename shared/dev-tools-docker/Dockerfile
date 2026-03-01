FROM golang:1.25-alpine

RUN apk add --no-cache \
    postgresql-client \
    sqlite \
    bash \
    curl

# Set PATH before installing Go binaries
ENV PATH="/go/bin:${PATH}"
ENV GOBIN=/go/bin

# Install migrate
RUN go install -tags 'postgres sqlite3' \
    github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Install sqlc
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Verify installations
RUN sqlc version && migrate -version

WORKDIR /app
