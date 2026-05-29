# BarterSwap — API d'échange de compétences

API REST en **Go** permettant à des particuliers d'échanger leurs compétences
sans argent, via un système de **crédit-temps** (1 heure rendue = 1 heure reçue).

> Projet de fin de module. Contraintes : stdlib uniquement (`net/http`,
> `encoding/json`, `database/sql`, `context`), **PostgreSQL** via le driver
> `github.com/lib/pq` (seule dépendance externe), **pas d'ORM**, **pas de
> framework**, **un seul package Go**.

## Installation

```bash
git clone <url>
cd barterswap

# 1. Démarrer PostgreSQL (Docker)
docker compose up -d db

# 2. Configurer l'environnement
cp .env.example .env        # adapter si besoin

# 3. Lancer l'API
go mod tidy
go run .
```

L'API écoute par défaut sur `http://localhost:8080`. Vérification rapide :

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

## Tests

```bash
go test -v -cover ./...
```
