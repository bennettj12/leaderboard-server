# Leaderboard API

A simple  leaderboard server written in Go. Games register to get an API key, clients submit scores, anyone can read leaderboards.

**Live:** [Leaderboard for 'A Game About Shooting Rocks'](https://leaderboard.bennett.click/api/leaderboard/2)
Game Link: [A Game About Shooting Rocks](https://bigghead.itch.io/rock-shooting-game)

## Authentication

| Header | Used for | Held by |
|---|---|---|
| `X-Admin-Key` | Creating / listing games | Operator (env `ADMIN_KEY`) |
| `X-API-Key` | Submitting scores | Each game (per-game key) |

## Endpoints

| Endpoint | Auth | Description |
|---|---|---|
| `GET /healthz` | — | Health check |
| `POST /api/games` | admin | Create a game. Returns the game incl. `api_key` — shown only once, store it |
| `GET /api/games` | admin | List all games (incl. `api_key`) |
| `POST /api/scores/{gameID}` | game key | Submit a score (see below). Best score per player is kept |
| `GET /api/leaderboard/{gameID}` | public | Top scores, highest first. `?limit=` (≤100, default 10), `?offset=` (default 0). `X-Total-Count` response header = total scores |
| `GET /api/leaderboard/{gameID}/{userID}` | public | A player's score & rank (`404` if none) |

## Submitting a score

```json
{"player_id": "<uuid>", "player_name": "Alice", "score": 12345, "hash": "<hex>"}
```

`hash` = lowercase hex SHA-256 of `player_name|player_id|score|api_key` — pipe-separated, score as a plain integer, `api_key` = the same value sent in the `X-API-Key` header. Server rejects mismatches with `400`.

Example: input `Alice|550e8400-e29b-41d4-a716-446655440000|12345|8f14e45f-ceea-4673-b9e2-8d9b3e1a2c4d` → hash `80531df864ab6b986c29c12314a8a50dce9eae8060449dab94b1cdb66c83eb45`

Code examples (Godot, JavaScript, Go): see [hash-examples.md](hash-examples.md).

The hash makes cheating harder but will not prevent cheating entirely as long as the score submit requests are sent from player's game client to the server.
