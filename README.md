# BarterSwap

Plateforme d'échange de compétences entre particuliers, basée sur un système de **crédit-temps** (1 heure rendue = 1 heure reçue, sans argent).

API REST écrite en **Go** (stdlib uniquement) avec **PostgreSQL**.

## Prérequis

- [Go](https://go.dev/dl/) 1.24+
- [Docker](https://www.docker.com/) (pour PostgreSQL)

## Installation

```bash
git clone <url>
cd BarterSwap
go mod tidy
```

## Configuration

Variables d'environnement (optionnelles) :

| Variable | Défaut | Description |
|----------|--------|-------------|
| `DATABASE_URL` | `postgres://barterswap:barterswap@localhost:5434/barterswap?sslmode=disable` | Connexion PostgreSQL |
| `PORT` | `8080` | Port d'écoute de l'API |

```bash
cp .env.example .env   # adapter si besoin
```

## Démarrage

```bash
# 1. Base de données
docker compose up -d db

# 2. API
go run .
```

L'API est disponible sur **http://localhost:8080**.

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

## Tests

```bash
go test -v -cover ./...
```

Collection Postman disponible dans `postman/` (import pour tester les endpoints).

## Documentation projet

Voir **[AGENTS.md](./AGENTS.md)** pour l'architecture, les conventions de code, l'état d'avancement et la roadmap.
