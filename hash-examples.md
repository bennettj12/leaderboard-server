# Score hash — code examples

Formula used by `POST /api/scores/{gameID}`:

```
hash = lowercase-hex( sha256( player_name | player_id | score | api_key ) )
```

Rules: pipe-separated, no spaces; `score` is a plain integer (`12345`, not `12345.0`); `api_key` is the same value sent in the `X-API-Key` header.

**Sample values** (all examples below should produce the same result):

```
input = Alice|550e8400-e29b-41d4-a716-446655440000|12345|8f14e45f-ceea-4673-b9e2-8d9b3e1a2c4d
hash  = 80531df864ab6b986c29c12314a8a50dce9eae8060449dab94b1cdb66c83eb45
```

---

## Godot (4.x)

```gdscript
func score_hash(player_name: String, player_id: String, score: int, api_key: String) -> String:
    var input := "%s|%s|%d|%s" % [player_name, player_id, score, api_key]
    return input.sha256_text()  # lowercase hex SHA-256 of the UTF-8 input
```

Godot 3 doesn't have `String.sha256_text()` — use `HashingContext` instead:

```gdscript
func score_hash(player_name: String, player_id: String, score: int, api_key: String) -> String:
    var ctx := HashingContext.new()
    ctx.start(HashingContext.HASH_SHA256)
    ctx.update(("%s|%s|%d|%s" % [player_name, player_id, score, api_key]).to_utf8())
    return ctx.finish().hex_encode()
```

## JavaScript

Node.js:

```js
const crypto = require('crypto');

function scoreHash(playerName, playerId, score, apiKey) {
  const input = `${playerName}|${playerId}|${score}|${apiKey}`;
  return crypto.createHash('sha256').update(input).digest('hex');
}
```

Browser (Web Crypto — async):

```js
async function scoreHash(playerName, playerId, score, apiKey) {
  const input = `${playerName}|${playerId}|${score}|${apiKey}`;
  const data = new TextEncoder().encode(input);
  const digest = await crypto.subtle.digest('SHA-256', data);
  return [...new Uint8Array(digest)].map(b => b.toString(16).padStart(2, '0')).join('');
}
```

> Note: in JS, `score` must be an integer (`Math.round`/`truncate` floats first, and don't exceed 2^53).

## Go

```go
import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func scoreHash(playerName, playerID string, score int64, apiKey string) string {
	sum := sha256.Sum256(fmt.Appendf(nil, "%s|%s|%d|%s", playerName, playerID, score, apiKey))
	return hex.EncodeToString(sum[:])
}
```
