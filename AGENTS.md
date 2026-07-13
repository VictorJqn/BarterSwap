# BarterSwap — Guide du projet

Document de référence pour les contributeurs et agents IA. Décrit les règles de code alignées sur le cours Go ESGI, l'architecture, l'état d'avancement et ce qu'il reste à faire.

---

## Contraintes du sujet (non négociables)

| Règle | Détail |
|-------|--------|
| Langage | Go uniquement |
| HTTP | `net/http` + `encoding/json` (stdlib) |
| Base de données | PostgreSQL via `github.com/lib/pq` — **seule dépendance externe** |
| Accès données | `database/sql` uniquement — **pas d'ORM** |
| Framework | **Aucun** (pas Gin, Echo, Chi…) |
| Packages | **Un seul package** `main` — pas de sous-packages internes |
| Concurrence | **Pas de mutex** — la base gère la concurrence |
| Auth | Header `X-User-ID` uniquement |

---

## Règles de clean code (cours + sujet)

### 1. Séparation des responsabilités

Le sujet l'exige explicitement : la logique métier ne doit **jamais** se mélanger avec HTTP ou SQL.

```
Handlers (HTTP)  →  Services (métier)  →  sqlStore (SQL)
```

| Couche | Fichiers | Autorisé | Interdit |
|--------|----------|----------|----------|
| Présentation | `handler_*.go`, `api.go`, `http.go`, `middleware.go` | Décoder JSON, appeler un service, écrire la réponse | Validations métier, requêtes SQL |
| Métier | `service_*.go` | Règles de gestion, autorisations, validations | `http.ResponseWriter`, SQL brut |
| Infrastructure | `store_*.go`, `db.go` | Requêtes SQL, transactions | Règles métier, codes HTTP |

### 2. Interfaces petites, côté consommateur (Module 6 du cours)

> *« BIEN : définir l'interface dans le package qui l'utilise »*

Chaque service déclare **son propre** contrat repository :

- `userRepository` dans `service_users.go`
- `serviceRepository` dans `service_services.go`
- `exchangeRepository` dans `service_exchanges.go`

`sqlStore` les implémente **implicitement** (duck typing). Pas de grosse interface `Store` centralisée.

```go
// Vérification à la compilation dans store.go
var (
    _ userRepository     = (*sqlStore)(nil)
    _ serviceRepository  = (*sqlStore)(nil)
    _ exchangeRepository = (*sqlStore)(nil)
)
```

### 3. Composition, pas d'héritage (Module 5)

- Structs + méthodes avec récepteur pointeur quand on modifie l'état
- Assembler les services dans `main.go`
- Pas de classes, pas d'héritage

### 4. Gestion d'erreurs idiomatique (Module 7)

```go
// Sentinelles dans errors.go
var ErrNotFound = errors.New("ressource introuvable")

// Wrapping avec contexte
return fmt.Errorf("%w: utilisateur %d", ErrNotFound, id)

// Comparaison en haut de la chaîne (handlers)
errors.Is(err, ErrNotFound)
```

Règles :
- Toujours vérifier `if err != nil`
- Ne pas ignorer les erreurs (`_` interdit sauf cas exceptionnel)
- Logger en haut de la chaîne (middleware / main), pas au milieu du store
- `panic` uniquement pour bugs — jamais comme gestion d'erreur normale
- Mapping HTTP centralisé dans `httpStatus()` (`errors.go`)

### 5. Conventions Go

- `go fmt` sur tout le code avant commit
- `go vet ./...` sans warnings
- Retours multiples `(valeur, error)` — pattern idiomatique
- Struct tags JSON sur les types du domaine (`models.go`)
- Noms exportés (majuscule) uniquement si nécessaire hors package — ici tout est `main`, donc peu pertinent
- `context.Context` en premier paramètre des fonctions store/service

### 6. Organisation des fichiers (mono-package)

Tous les `.go` sont à la **racine** du module. C'est normal et imposé par le sujet.

| Fichier | Rôle |
|---------|------|
| `main.go` | Point d'entrée, câblage des dépendances |
| `config.go` | Variables d'environnement |
| `models.go` | Types du domaine + constantes + validateurs |
| `errors.go` | Erreurs sentinelles + mapping HTTP |
| `db.go` | Connexion PostgreSQL + schéma (`migrate`) |
| `store.go` | Struct `sqlStore` + constructeur |
| `store_users.go` | SQL utilisateurs + skills |
| `store_services.go` | SQL services |
| `store_exchanges.go` | SQL échanges + journal crédits |
| `service_users.go` | Métier utilisateurs |
| `service_services.go` | Métier services |
| `service_exchanges.go` | Métier échanges |
| `api.go` | Enregistrement des routes |
| `handler_users.go` | Handlers HTTP utilisateurs |
| `handler_services.go` | Handlers HTTP services |
| `handler_exchanges.go` | Handlers HTTP échanges |
| `http.go` | Helpers JSON, parsing IDs |
| `middleware.go` | Logging, recovery, CORS, auth |

### 7. Base de données

- Schéma dans `db.go` (const `schema`), appliqué au démarrage
- Solde crédits = `SUM(montant)` sur `credit_transactions` (journal, pas de colonne solde)
- Types de transaction : `earn`, `spend`, `refund`
- Index unique `uq_exchanges_active_service` : un seul échange `pending`/`accepted` par service

### 8. Tests (objectif sujet : ≥ 60 % couverture)

- Tests **table-driven** (`testing` + sous-tests)
- Tests **unitaires** des services avec mock repository (interface = facile à mocker)
- Tests **API** avec `httptest` + base de test
- Chaque cas métier du sujet doit avoir au moins un test

---

## Schéma de la base

```
users ──< skills
users ──< services (provider_id)
users ──< exchanges (requester_id, owner_id)
users ──< credit_transactions
users ──< reviews (author_id, target_id)
services ──< exchanges
exchanges ──< credit_transactions
exchanges ──< reviews
```

Tables : `users`, `skills`, `services`, `exchanges`, `credit_transactions`, `reviews`

---

## Endpoints

### Implémentés

| Méthode | Path | Auth |
|---------|------|------|
| GET | `/health` | — |
| POST | `/api/users` | — |
| GET | `/api/users/{id}` | — |
| PUT | `/api/users/{id}` | ✓ |
| GET | `/api/users/{id}/skills` | — |
| PUT | `/api/users/{id}/skills` | ✓ |
| GET | `/api/services` | — |
| POST | `/api/services` | ✓ |
| GET | `/api/services/{id}` | — |
| PUT | `/api/services/{id}` | ✓ |
| DELETE | `/api/services/{id}` | ✓ |
| POST | `/api/exchanges` | ✓ |
| GET | `/api/exchanges` | ✓ |
| GET | `/api/exchanges/{id}` | ✓ |
| PUT | `/api/exchanges/{id}/accept` | ✓ |
| PUT | `/api/exchanges/{id}/reject` | ✓ |
| PUT | `/api/exchanges/{id}/complete` | ✓ |
| PUT | `/api/exchanges/{id}/cancel` | ✓ |

### À implémenter

| Méthode | Path | Auth | Section sujet |
|---------|------|------|---------------|
| POST | `/api/exchanges/{id}/review` | ✓ | Évaluations |
| GET | `/api/users/{id}/reviews` | — | Évaluations |
| GET | `/api/services/{id}/reviews` | — | Évaluations |
| GET | `/api/users/{id}/stats` | — | Statistiques |

---

## Ce qui est fait

### Infrastructure
- [x] Module Go + `go.mod`
- [x] Docker Compose PostgreSQL (port 5434)
- [x] Configuration via variables d'environnement
- [x] Migration automatique du schéma au démarrage
- [x] Middlewares : logging, recovery, CORS

### Utilisateurs
- [x] Création avec 10 crédits de bienvenue (journal `credit_transactions`)
- [x] Profil public, modification profil (auth)
- [x] Compétences : lecture + remplacement complet (PUT écrase)
- [x] Validation pseudo obligatoire, niveaux de compétence

### Services
- [x] CRUD complet (auth pour écriture)
- [x] Filtres serveur : `categorie`, `ville`, `search`
- [x] Vérification que l'utilisateur possède la compétence (= catégorie)
- [x] Catégories fermées (liste du sujet)

### Échanges
- [x] Création sur un `service_id` (owner déduit du provider)
- [x] Cycle de vie : pending → accepted → completed (+ rejected, cancelled)
- [x] Pas d'échange sur son propre service
- [x] Vérification solde avant création
- [x] Un seul échange actif par service (409)
- [x] Accept : débit (`spend`) du demandeur
- [x] Complete : crédit (`earn`) à l'offreur
- [x] Cancel/Reject : remboursement (`refund`) si crédits bloqués
- [x] Liste filtrable par `status`

### Outils
- [x] Collection Postman (`postman/`) avec démo complète + cas d'erreur

---

## Ce qu'il reste à faire

### Fonctionnalités (sujet)

- [ ] **Évaluations** (`POST /api/exchanges/{id}/review`)
  - Un seul avis par utilisateur par échange
  - Uniquement sur échange `completed`
  - Note 1–5, pas de modification/suppression
  - `GET /api/users/{id}/reviews` et `GET /api/services/{id}/reviews`

- [ ] **Statistiques** (`GET /api/users/{id}/stats`)
  - `services_actifs`, `echanges_completes`, `credit_balance`
  - `note_moyenne`, `nb_avis`, `total_gagne`, `total_depense`

### Tests (critère : 3 pts, ≥ 60 % couverture)

- [ ] Tests unitaires services (table-driven)
- [ ] Tests API avec `httptest`
- [ ] Cas métier du sujet couverts :
  1. Créer utilisateur → 201
  2. Pseudo vide → 400
  3. Service sans compétence → 400
  4. Échange sur son propre service → 400
  5. Crédits insuffisants → 400
  6. Service déjà réservé → 409
  7. Accept → crédits bloqués
  8. Complete → crédits transférés
  9. Cancel → crédits restitués
  10. Noter échange non terminé → 400
  11. Noter deux fois → 400
  12. Stats cohérentes

### Documentation & démo

- [ ] README : exemples curl (optionnel, Postman suffit)
- [ ] Script `demo.sh` backup curl pour la soutenance
- [ ] Commentaires godoc sur les fonctions exportées / publiques clés

### Qualité

- [ ] `go test -cover` ≥ 60 %
- [ ] `go vet` clean
- [ ] Vérifier tous les codes HTTP sur les cas d'erreur

---

## Règles métier — référence rapide

### Crédits
| Événement | Transaction | Montant |
|-----------|-------------|---------|
| Création compte | `earn` | +10 |
| Accept échange | `spend` | −credits (demandeur) |
| Complete échange | `earn` | +credits (offreur) |
| Cancel/Reject (si accepted) | `refund` | +credits (demandeur) |

### Échange
- On demande un échange via `service_id`, pas via `user_id`
- `owner_id` = `provider_id` du service
- `requester_id` = utilisateur connecté (`X-User-ID`)

### Autorisations
| Action | Qui |
|--------|-----|
| Modifier profil/skills | Soi-même uniquement |
| CRUD service | Provider uniquement |
| Créer échange | N'importe qui (sauf sur son propre service) |
| Accept / Reject / Complete | Owner (offreur) |
| Cancel | Requester ou Owner |

---

## Critères d'évaluation (soutenance)

| Critère | Points | État |
|---------|--------|------|
| Fonctionnalités | /5 | ~3/5 (manque reviews + stats) |
| Architecture | /3 | Bonne séparation en couches |
| Tests | /3 | 0/3 (pas encore écrits) |
| Qualité code | /2 | gofmt + go vet OK |
| Documentation | /1 | README + ce fichier |
| Gestion erreurs | /1 | Sentinelles + wrapping en place |
| Bonus jury | /5 | Postman collection |

---

## Commandes utiles

```bash
# Développement
docker compose up -d db
go run .

# Qualité
go fmt ./...
go vet ./...
go test -v -cover ./...

# Reset base
docker compose down -v && docker compose up -d db

# Inspecter la base
docker compose exec db psql -U barterswap -d barterswap -c "\dt"
```
