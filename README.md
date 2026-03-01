# eev

> Share text, URLs, and files with a 6-digit code. No email drafts, no faff.

---

<!-- Screenshot or demo image -->
![Share flow](docs/images/share.png)

---

## Why does this exist

Picture this. Exams tomorrow. Notes are on the iPad. Need them on the university workstation. AirDrop? Nope, not on the university machine. iCloud? Locked down. So what do I do? Open my email, paste my notes into a **draft**, walk over to the other machine, and copy it out like it's 2003.

That was the moment. The name eev came from that evening!

## The problem (aka the email draft hack)

You want to send a link or a bit of text from one device to another. Simple, right? Turns out if the two devices aren't in the same ecosystem, you're basically on your own. The go-to solution for most people — including, embarrassingly, me for years — is the email draft. Type it in, don't send it, go to the other device, fish it out. Every. Single. Time.

There had to be a better way. So I built it.

## How it works

1. Share something — a URL, a snippet of text, or a file
2. Get a **6-digit code**
3. On any other device, go to `e.alenalex.me` and enter the code
4. If it's a URL, you're taken straight there. If it's text or a file, it's right there waiting.

That's it.

## Clients

| Client | Status | Use case |
|--------|--------|----------|
| Web | Planned | Any device, any browser |
| Apple Shortcut | Planned | Share straight from iOS/macOS share sheet |
| TUI | Planned | Quick access from the terminal |

## Monorepo Structure

```
eev/
├── backend/   # Go API server
└── frontend/  # Web client
```

- [Backend →](./apps/backend-go/README.md)
- [Frontend →](./apps/eev-web/README.md)

## Self-hosted

eev is built for personal use. You run your own instance, your data stays yours.

Everything — frontend, backend, and reverse proxy — ships as a single Docker image. No compose file, no separate services.

```bash
docker run -d \
  -p 8080:80 \
  -e S3_ACCESS_KEY=... \
  -e S3_SECRET_KEY=... \
  -e S3_BUCKET=... \
  .
  . //WILL COMPLETE IT AFTER THE PROJECT IS DONE
  -v eev_data:/data \
  ghcr.io/alenalex/eev
```

The SQLite database is persisted to `/data` — mount a volume to keep it across restarts.
