# BarterSwap

Plateforme d'échange de compétences entre particuliers, basée sur un système de **crédit-temps** (1 heure rendue = 1 heure reçue, sans argent).

API REST en **Go** (stdlib uniquement) avec **PostgreSQL**.

## Prérequis

- [Go](https://go.dev/dl/) 1.24+
- [Docker](https://www.docker.com/) (PostgreSQL)

## Installation

```bash
git clone <url>
cd BarterSwap
go mod tidy
cp .env.example .env   # optionnel
```

## Configuration

| Variable | Défaut | Description |
|----------|--------|-------------|
| `DATABASE_URL` | `postgres://barterswap:barterswap@localhost:5434/barterswap?sslmode=disable` | Connexion PostgreSQL |
| `PORT` | `8080` | Port d'écoute |

## Démarrage

```bash
docker compose up -d db
go run .
```

L'API est disponible sur **http://localhost:8080**.

## Authentification

Pas de JWT : le header `X-User-ID` identifie l'utilisateur connecté sur les routes protégées.

## Documentation API

La liste complète des endpoints, schémas et codes de réponse est disponible dans la spec **OpenAPI (Swagger)** :

- [`docs/openapi.yaml`](docs/openapi.yaml)

Pour la visualiser : [editor.swagger.io](https://editor.swagger.io) → Import → `docs/openapi.yaml`

### Postman

Des fichiers JSON prêts à l'import sont disponibles dans `postman/` :

| Fichier | Description |
|---------|-------------|
| `postman/BarterSwap.postman_collection.json` | Collection complète (démo, endpoints, cas d'erreur) |
| `postman/BarterSwap.postman_environment.json` | Variables d'environnement (`base_url`, `user_id`, etc.) |

Dans Postman : **Import** → sélectionner les deux fichiers → choisir l'environnement **BarterSwap — Local**.

## Exemples d'utilisation

Scénario nominal avec `curl` (API sur `http://localhost:8080`) :

```bash
BASE=http://localhost:8080

# 1. Vérifier que l'API répond
curl -s $BASE/health

# 2. Créer deux utilisateurs (10 crédits de bienvenue chacun)
curl -s -X POST $BASE/api/users \
  -H 'Content-Type: application/json' \
  -d '{"pseudo":"alice","ville":"Paris"}'

curl -s -X POST $BASE/api/users \
  -H 'Content-Type: application/json' \
  -d '{"pseudo":"bob","ville":"Lyon"}'

# 3. Alice déclare une compétence et publie un service
curl -s -X PUT $BASE/api/users/1/skills \
  -H 'Content-Type: application/json' -H 'X-User-ID: 1' \
  -d '[{"nom":"Jardinage","niveau":"expert"}]'

curl -s -X POST $BASE/api/services \
  -H 'Content-Type: application/json' -H 'X-User-ID: 1' \
  -d '{"titre":"Tonte de pelouse","categorie":"Jardinage","duree_minutes":60,"credits":2,"ville":"Paris"}'

# 4. Bob demande un échange, Alice accepte et termine
curl -s -X POST $BASE/api/exchanges \
  -H 'Content-Type: application/json' -H 'X-User-ID: 2' \
  -d '{"service_id":1}'

curl -s -X PUT $BASE/api/exchanges/1/accept -H 'X-User-ID: 1'
curl -s -X PUT $BASE/api/exchanges/1/complete -H 'X-User-ID: 1'

# 5. Bob note Alice, puis consulte ses statistiques
curl -s -X POST $BASE/api/exchanges/1/review \
  -H 'Content-Type: application/json' -H 'X-User-ID: 2' \
  -d '{"note":5,"commentaire":"Excellent travail"}'

curl -s $BASE/api/users/1/stats
```

Cas d'erreur (pseudo vide → 400) :

```bash
curl -s -o /dev/null -w "%{http_code}\n" -X POST $BASE/api/users \
  -H 'Content-Type: application/json' \
  -d '{"pseudo":""}'
```

## Architecture

Mono-package `main` avec séparation en couches :

```
Handlers (HTTP)  →  Services (métier)  →  sqlStore (SQL)
```

| Couche | Fichiers |
|--------|----------|
| Présentation | `handler_*.go`, `api.go`, `middleware.go` |
| Métier | `service_*.go` |
| Infrastructure | `store_*.go`, `db.go` |

Les crédits sont gérés via un **journal** (`credit_transactions`) : `earn`, `spend`, `refund`.

### Godoc

Le code est documenté avec des commentaires godoc sur le package, les types exportés, les services et les erreurs sentinelles.

```bash
# Vue d'ensemble du package
go doc .

# Documentation complète (types, constantes, fonctions)
go doc -all .

# Un service métier et ses méthodes
go doc UserService
go doc -all UserService

# Un type du domaine
go doc User
go doc Exchange

# Les erreurs et leur mapping HTTP
go doc ErrValidation
go doc httpStatus

# Lancer un serveur godoc local (optionnel)
go install golang.org/x/tools/cmd/godoc@latest
godoc -http=:6060
# puis ouvrir http://localhost:6060/pkg/barterswap/
```

## Tests

```bash
docker compose up -d db
go test -v -cover ./...
```

Sans Docker, les tests unitaires et API passent ; les tests d'intégration sont skippés. Couverture actuelle : **~74 %**.

| Fichier | Type |
|---------|------|
| `service_*_test.go` | Unitaires (table-driven) |
| `api_test.go` | API (`httptest`) |
| `integration_test.go` | Intégration PostgreSQL |
| `mock_test.go` | Mock en mémoire |
| `config_test.go`, `http_test.go`, `middleware_test.go` | Utilitaires et middleware |
