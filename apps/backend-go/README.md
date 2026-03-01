# eev / backend

The API server for eev. Built in Go — this is my first Go project, so keep that in mind.

## What it does

- Auth
- Creating shareables (text, URLs, files)
- Viewing/retrieving shareables by 6-digit code
- File storage via S3-compatible backend

## Stack

- **Go**
- **SQLite** — simple, no infra overhead
- **S3-compatible storage** — for file shareables (Cloudflare R2, AWS S3, or LocalStack locally)

## Local Development

Start a local S3 stack:

```bash
docker compose up -d
./localstack-bucket.sh create eev
```

Copy and configure your environment:

```bash
cp .env.example .env
```

Run the server:

```bash
go run ./cmd/server
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `S3_ACCESS_KEY` | S3 access key | |
| `S3_SECRET_KEY` | S3 secret key | |
| `S3_REGION` | S3 region | `auto` |
| `S3_BUCKET` | S3 bucket name | |
| `S3_ENDPOINT_URL` | Custom S3 endpoint (for R2, LocalStack) | |