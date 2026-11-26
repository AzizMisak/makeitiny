// Router PostgreSQL fallback'i: tıklama sınırı denetimine henüz aktarılmamış Redis farkları da katılır.
const fs = require('fs')
let miss = 0
function edit(p, pairs) {
  let s = fs.readFileSync(p, 'utf8')
  for (const [a, b] of pairs) {
    if (!s.includes(a)) { console.log('MISS', p, JSON.stringify(a.slice(0, 80))); miss++; continue }
    s = s.split(a).join(b)
  }
  fs.writeFileSync(p, s)
}
const base = 'Backend/router/internal/'
edit(base + 'db/queries.sql', [
  [`-- name: IncrementLinkClicks :execrows
UPDATE links
SET clicks = clicks + 1
WHERE alias = $1
  AND (max_clicks IS NULL OR max_clicks <= 0 OR clicks < max_clicks);`, `-- name: IncrementLinkClicks :execrows
UPDATE links
SET clicks = clicks + 1
WHERE alias = $1
  AND (max_clicks IS NULL OR max_clicks <= 0 OR clicks + $2 < max_clicks);`]
])
edit(base + 'db/queries.sql.go', [
  [`const incrementLinkClicks = \`-- name: IncrementLinkClicks :execrows
UPDATE links
SET clicks = clicks + 1
WHERE alias = $1
  AND (max_clicks IS NULL OR max_clicks <= 0 OR clicks < max_clicks)
\`

func (q *Queries) IncrementLinkClicks(ctx context.Context, alias string) (int64, error) {
	res, err := q.db.Exec(ctx, incrementLinkClicks, alias)`, `const incrementLinkClicks = \`-- name: IncrementLinkClicks :execrows
UPDATE links
SET clicks = clicks + 1
WHERE alias = $1
  AND (max_clicks IS NULL OR max_clicks <= 0 OR clicks + $2 < max_clicks)
\`

type IncrementLinkClicksParams struct {
	Alias  string \`json:"alias"\`
	Clicks int64  \`json:"clicks"\`
}

func (q *Queries) IncrementLinkClicks(ctx context.Context, arg IncrementLinkClicksParams) (int64, error) {
	res, err := q.db.Exec(ctx, incrementLinkClicks, arg.Alias, arg.Clicks)`]
])
edit(base + 'repository/repository.go', [
  [`	// Fallback: Redis devre dışıysa veya sınırlı link anahtarı eksikse doğrudan PostgreSQL'e yaz
	rows, err := r.q.IncrementLinkClicks(ctx, rec.Alias)`, `	// Fallback: Redis devre dışıysa veya sınırlı link anahtarı eksikse doğrudan PostgreSQL'e yaz.
	// Sınır denetimine henüz worker ile aktarılmamış tıklamalar da katılır; yoksa sınır aşılabilirdi.
	var pending int64
	if rec.MaxClicks > 0 {
		pending, _ = r.pendingClicks(ctx, rec.ID, rec.Alias)
	}
	rows, err := r.q.IncrementLinkClicks(ctx, db.IncrementLinkClicksParams{Alias: rec.Alias, Clicks: pending})`]
])
console.log(miss ? 'misses ' + miss : 'fix2 ok')
