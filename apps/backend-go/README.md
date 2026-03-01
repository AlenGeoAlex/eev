# eev

A personal content sharing tool that supports text, URLs, and files — not just files like everything else.

## Why

Sharing a URL or a piece of text across devices usually means dumping it into an email draft and copying it over. eev cuts that out entirely.

Share something, get a 6-digit code. On any other device, go to `e.alenalex.me`, enter the code — done. If it's a URL, it takes you straight there.

## Clients

| Client | Status | Use case |
|--------|--------|----------|
| Web | Planned | Universal, works on any device |
| Apple Shortcut | Planned | Quick share from iPhone/iPad/Mac |
| TUI | Planned | Terminal-first workflow on laptops |

## What you can share

- **Text** — snippets, notes, anything
- **URLs** — links you want on another device
- **Files** — when you actually need to move a file

## Self-hosted

eev is built for personal use. You run your own instance, your data stays yours.

## Stack

- **Backend** — Go
- **Storage** — S3-compatible (works with Cloudflare R2, LocalStack, AWS S3)

## Local Development

Start a local S3-compatible stack:

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