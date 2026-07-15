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
go run ./cmd/barterswap
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

## Structure du projet

```
BarterSwap/
├── cmd/
│   └── barterswap/
│       └── main.go                 # Point d'entrée, câblage des dépendances
├── internal/
│   ├── domain/
│   │   ├── doc.go                  # Documentation du package
│   │   └── models.go               # User, Service, Exchange, Review, UserStats…
│   ├── apperr/
│   │   └── errors.go               # Erreurs sentinelles + mapping HTTP
│   ├── config/
│   │   └── config.go               # Variables d'environnement (DATABASE_URL, PORT)
│   ├── database/
│   │   └── db.go                   # Connexion PostgreSQL + schéma SQL
│   ├── repository/
│   │   ├── store.go                # Constructeur Store
│   │   ├── store_users.go          # SQL utilisateurs + compétences
│   │   ├── store_services.go       # SQL annonces
│   │   ├── store_exchanges.go      # SQL échanges + journal crédits
│   │   ├── store_reviews.go        # SQL avis
│   │   └── store_stats.go          # SQL statistiques
│   ├── service/
│   │   ├── users.go                # Logique métier utilisateurs
│   │   ├── services.go             # Logique métier annonces
│   │   ├── exchanges.go            # Logique métier échanges
│   │   └── reviews.go              # Logique métier avis
│   └── httpapi/
│       ├── api.go                  # Enregistrement des routes
│       ├── handler_users.go        # Handlers utilisateurs
│       ├── handler_services.go     # Handlers services
│       ├── handler_exchanges.go    # Handlers échanges
│       ├── handler_reviews.go      # Handlers avis
│       ├── http.go                 # Helpers JSON, parsing IDs
│       └── middleware.go           # Logging, recovery, CORS, auth
├── tests/
│   ├── mock/
│   │   └── store.go                # Store en mémoire (tests unitaires + API)
│   ├── unit/
│   │   ├── users_test.go           # Tests UserService
│   │   ├── services_test.go        # Tests ServiceService
│   │   ├── exchanges_test.go       # Tests ExchangeService
│   │   ├── reviews_test.go         # Tests ReviewService
│   │   ├── errors_test.go          # Tests mapping HTTP
│   │   ├── config_test.go          # Tests configuration
│   │   └── http_test.go            # Tests middleware / JSON
│   ├── api/
│   │   ├── helpers.go              # Helper newTestAPI
│   │   └── api_test.go             # Tests httptest (endpoints)
│   └── integration/
│       ├── flows_test.go           # Scénarios complets PostgreSQL
│       └── db_test.go              # Tests connexion / schéma
├── docs/
│   └── openapi.yaml                # Spec OpenAPI (Swagger)
├── postman/
│   ├── BarterSwap.postman_collection.json
│   └── BarterSwap.postman_environment.json
├── docker-compose.yml              # PostgreSQL local
├── go.mod
└── README.md
```

## Architecture

Séparation en couches (clean architecture) :

```
httpapi (handlers)  →  service (métier)  →  repository (SQL)
```

| Couche | Package | Rôle |
|--------|---------|------|
| Présentation | `internal/httpapi` | Routes, JSON, middleware, auth header |
| Métier | `internal/service` | Règles de gestion, validations, cycle de vie |
| Infrastructure | `internal/repository` | Requêtes SQL, transactions, journal de crédits |
| Domaine | `internal/domain` | Types partagés (`User`, `Exchange`, etc.) |

Les crédits sont gérés via un **journal** (`credit_transactions`) : `earn`, `spend`, `refund`.

### Godoc

```bash
# Service métier utilisateurs
go doc barterswap/internal/service UserService

# Types du domaine
go doc barterswap/internal/domain User

# Erreurs et codes HTTP
go doc barterswap/internal/apperr ErrValidation
go doc barterswap/internal/apperr HTTPStatus

# Handler HTTP
go doc barterswap/internal/httpapi API

# Serveur godoc local (optionnel)
go install golang.org/x/tools/cmd/godoc@latest
godoc -http=:6060
```

## Tests

```bash
docker compose up -d db
go test -v -cover ./...
```

Sans Docker, les tests unitaires et API passent ; les tests d'intégration sont skippés.

| Dossier | Type |
|---------|------|
| `tests/unit/` | Unitaires (table-driven) |
| `tests/api/` | API (`httptest`) |
| `tests/integration/` | Intégration PostgreSQL |
| `tests/mock/` | Mock en mémoire |
