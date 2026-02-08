# Krizzy

Lightweight Kanban board built with Go, Echo, Templ, Datastar, and SQLite.

![board](./board.png)
![modal](./modal.png)

## Requirements

- **Go 1.21+** - Required for building and running the app
- **Node.js + npm** - Optional, for building Tailwind CSS locally (uses CDN by default)

## Run

```bash
make run
```

Open http://localhost:8080

## Build

```bash
make build
./bin/krizzy
```

## Postgres (optional)

Boards can optionally be backed by Postgres instead of SQLite. A Docker container is included for local development (requires Docker and may need `sudo`):

```bash
make pg-up      # start Postgres in background
make pg-down    # stop Postgres
make pg-reset   # stop Postgres and wipe data
```

Once Postgres is running, click **Manage Connections** on the boards page and add a connection with host=`localhost`, port=`5432`, user=`krizzy`, password=`krizzy`. Then create a new board and select "PostgreSQL" as the database type.

## Configuration

| Env Variable | Default | Description |
|--------------|---------|-------------|
| `SERVER_ADDRESS` | `:8080` | Server listen address |
| `DATABASE_PATH` | `krizzy.db` | SQLite database path |
