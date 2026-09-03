# Leaderboard API

A simple leaderboard server in Go. Games register once to receive an API key; clients submit scores under a `player_id`, and leaderboards are read back publicly with pagination.

**Live instance:** `https://leaderboard.bennett.click`

---

## Table of contents

- [Authentication](#authentication)
- [Endpoints](#endpoints)
  - [Health check](#health-check)
  - [Create a game](#create-a-game--admin-)
  - [List games](#list-games--admin-)
  - [Submit a score](#submit-a-score)
  - [Get leaderboard](#get-leaderboard)
  - [Get a player's score](#get-a-players-score)
- [Data model & semantics](#data-model--semantics)
- [Rate limiting & limits](#rate-limiting--limits)
- [Error codes](#error-codes)
- [Configuration](#configuration)
- [Local development](#local-development)

---

## Authentication

There are **two distinct keys**. Keep both secret.

| Header | Used for | Who holds it | Where it comes from |
|---|---|---|---|
| `X-Admin-Key` | Creating & listing games | Server operator only | `ADMIN_KEY` env var / `.env` |
| `X-API-Key` | Submitting scores | Game client | Returned once by `POST /api/games` |

- Admin endpoints return `401 Unauthorized` if the `X-Admin-Key` header is missing or wrong.
- Score submissions require the game's `X-API-Key`, and the key must belong to the game named in the URL path.

---

## Endpoints

### Health check

```
GET /healthz
```

No auth. Returns `200 OK` with an empty body. Use it for uptime checks / container healthchecks.

---

### Create a game  *(admin)*

```
POST /api/games
X-Admin-Key: <admin key>
Content-Type: application/json
```

**Request body:**

```json
{ "name": "My Game" }
```

**Response — `201 Created`:**

```json
{
  "id": 1,
  "name": "My Game",
  "api_key": "8f14e45f-ceea-4673-b9e2-8d9b3e1a2c4d",
  "created_at": "2026-09-03T15:28:30Z"
}
```

> ⚠️ `api_key` is returned **only here**. Store it — this is the key clients will send as `X-API-Key` when submitting scores.

**Errors:** `400` invalid JSON / empty name · `401` bad or missing admin key · `500` write failure (including a duplicate game name).

```bash
curl -X POST https://leaderboard.bennett.click/api/games \
  -H "Content-Type: application/json" \
  -H "X-Admin-Key: $ADMIN_KEY" \
  -d '{"name":"My Game"}'
```

---

### List games  *(admin)*

```
GET /api/games
X-Admin-Key: <admin key>
```

Returns **all** games as a JSON array (full objects, including `api_key` — useful for recovering/redistributing keys). Empty when there are none.

**Response — `200 OK`:**

```json
[
  {
    "id": 1,
    "name": "My Game",
    "api_key": "8f14e45f-ceea-4673-b9e2-8d9b3e1a2c4d",
    "created_at": "2026-09-03T15:28:30Z"
  }
]
```

**Errors:** `401` bad or missing admin key · `500` database failure.

---

### Submit a score

```
POST /api/scores/{gameID}
X-API-Key: <that game's api key>
Content-Type: application/json
```

**Request body:**

```json
{
  "player_id": "550e8400-e29b-41d4-a716-446655440000",
  "player_name": "Alice",
  "score": 12345
}
```

| Field | Rules |
|---|---|
| `player_id` | Required. Must be a valid UUID. **This is the rate-limit key.** |
| `player_name` | Required. Non-empty, ≤ `max_name_length` (24 by default). |
| `score` | Required. `0 ≤ score ≤ max_score` (10,000,000 by default). |

Only the player's **best** score is kept: submitting a lower score for an existing `player_id` is accepted as a no-op with `accepted: false`.

**Response — `200 OK`:**

```json
{
  "accepted": true,
  "message": "Score accepted",
  "current_best": 0,
  "submitted_score": 12345
}
```

`current_best` is the player's best score *before* this submission (`0` if they had none). When a higher score already exists: `accepted: false`, `message: "Higher score already exists"`.

**Errors:** `400` invalid JSON / invalid request (bad UUID, empty name, score out of range) · `401` missing or wrong `X-API-Key` · `413` body too large · `429` rate limited · `500` save failure.

```bash
curl -X POST https://leaderboard.bennett.click/api/scores/1 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $GAME_API_KEY" \
  -d '{"player_id":"550e8400-e29b-41d4-a716-446655440000","player_name":"Alice","score":12345}'
```

> Note: submitting to a `gameID` that doesn't exist currently returns `500` (not `404`). Create the game first.

---

### Get leaderboard

```
GET /api/leaderboard/{gameID}
```

Public — no auth. Returns the top scores for a game, **highest score first**.

**Query parameters:**

| Param | Default | Notes |
|---|---|---|
| `limit` | `10` | 1–100 entries per page |
| `offset` | `0` | Number of entries to skip (page N → `offset = (N-1) * limit`) |

**Response — `200 OK`.** A JSON array of entries plus an `X-Total-Count` **header** with the total number of scores for that game (not just the page size — use it to build pager controls):

```json
[
  {
    "player_name": "Alice",
    "score": 12345,
    "created_at": "2026-09-03T15:28:31Z",
    "rank": 1
  }
]
```

```bash
curl https://leaderboard.bennett.click/api/leaderboard/1?limit=20&offset=40
```

**Errors:** `400` missing game ID / invalid `limit` / invalid `offset` · `500` database failure.

---

### Get a player's score

```
GET /api/leaderboard/{gameID}/{userID}
```

Public — no auth. Returns one player's score and rank within that game.

**Response — `200 OK`:**

```json
{
  "player_name": "Alice",
  "score": 12345,
  "created_at": "2026-09-03T15:28:31Z",
  "rank": 1
}
```

**Errors:** `404` no score found for that player in that game · `500` database failure.

```bash
curl https://leaderboard.bennett.click/api/leaderboard/1/550e8400-e29b-41d4-a716-446655440000
```

---

## Data model & semantics

**Ranking** uses `RANK()`: players with equal scores **share a rank**, and the next rank is skipped (e.g. two players tied at #1 means the next player is #3). Tied rows are ordered deterministically by `player_id`.

**Persistence** — SQLite (`leaderboard.db`). A player has at most one row per game (`game_id` + `player_id` is unique); only their best score is retained.

**Timestamps** are RFC 3339 (e.g. `2026-09-03T15:28:31Z`).

---

## Rate limiting & limits

Only `POST /api/scores/{gameID}` is rate limited:

- Fixed **1-minute window** per `player_id`.
- Up to `max_requests` (30 by default) submissions per minute.
- Exceeding the limit returns `429` with body `rate limited`.

Request bodies are capped at `max_body_size` (1 MiB by default); larger bodies are rejected with `413`.

---

## Error codes

| Code | Meaning |
|---|---|
| `400` | Malformed JSON, missing/invalid field, bad `limit`/`offset` |
| `401` | Missing or invalid `X-Admin-Key` / `X-API-Key` |
| `404` | Player/game not found (score lookups) |
| `413` | Request body exceeds `max_body_size` |
| `429` | Rate limit exceeded |
| `500` | Server/database error |

Error responses are plain text messages describing the problem.

---

## Configuration

**`config.json`**

| Key | Default | Meaning |
|---|---|---|
| `max_score` | `10000000` | Upper bound for submitted scores |
| `max_requests` | `30` | Rate limit: submissions per player per minute |
| `max_name_length` | `24` | Max player name length |
| `port` | `8000` | HTTP listen port |
| `max_body_size` | `1048576` | Max request body size in bytes |

**Environment / `.env`**
| Key | Required | Meaning |
|---|---|---|
| `ADMIN_KEY` | Yes | Admin key for `X-Admin-Key` auth (startup fails if unset) |
| `DB_NAME` | No | SQLite file name (`leaderboard` → `leaderboard.db`) |
| `TEST_DB_NAME` | No | SQLite file name used by tests |

---

## Local development

```bash
# requires Go 1.27+
cp config_example.json config.json   # if config.json isn't present
echo "ADMIN_KEY=dev-admin-key" > .env

go run .        # starts on :8000
go test ./...   # runs the test suite (uses TEST_DB_NAME)
```

### Docker

```bash
docker compose up -d --build
curl http://127.0.0.1:8000/healthz
docker compose down     # stop (keeps the database volume)
docker compose down -v  # stop and wipe the database volume
```

The container binds to `127.0.0.1:8000` and expects a reverse proxy (nginx) in front of it for public TLS access.
