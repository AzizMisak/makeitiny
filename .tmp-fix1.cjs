// Çıkış, parola değişimi, askı ve tekrar kullanım tespitinde refresh token'lar silinir (iptal işareti yalnızca rotasyonda).
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
const base = 'Backend/api/internal/'

edit(base + 'db/queries.sql', [
  [`-- name: RevokeRefreshTokenByHash :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE token_hash = $1 AND revoked_at IS NULL;

-- name: RevokeAllUserRefreshTokens :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = $1 AND revoked_at IS NULL;`, `-- name: DeleteRefreshTokenByHash :exec
DELETE FROM refresh_tokens
WHERE token_hash = $1;

-- name: DeleteUserRefreshTokens :exec
DELETE FROM refresh_tokens
WHERE user_id = $1;`]
])

edit(base + 'db/queries.sql.go', [
  [`const revokeAllUserRefreshTokens = \`-- name: RevokeAllUserRefreshTokens :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = $1 AND revoked_at IS NULL
\`

func (q *Queries) RevokeAllUserRefreshTokens(ctx context.Context, userID int64) error {
	_, err := q.db.Exec(ctx, revokeAllUserRefreshTokens, userID)
	return err
}`, ''],
  [`const revokeRefreshTokenByHash = \`-- name: RevokeRefreshTokenByHash :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE token_hash = $1 AND revoked_at IS NULL
\`

func (q *Queries) RevokeRefreshTokenByHash(ctx context.Context, tokenHash string) error {
	_, err := q.db.Exec(ctx, revokeRefreshTokenByHash, tokenHash)
	return err
}`, ''],
  ['const deleteUser = ', `const deleteRefreshTokenByHash = \`-- name: DeleteRefreshTokenByHash :exec
DELETE FROM refresh_tokens
WHERE token_hash = $1
\`

func (q *Queries) DeleteRefreshTokenByHash(ctx context.Context, tokenHash string) error {
	_, err := q.db.Exec(ctx, deleteRefreshTokenByHash, tokenHash)
	return err
}

const deleteUser = `],
  ['const getLinkByAlias = ', `const deleteUserRefreshTokens = \`-- name: DeleteUserRefreshTokens :exec
DELETE FROM refresh_tokens
WHERE user_id = $1
\`

func (q *Queries) DeleteUserRefreshTokens(ctx context.Context, userID int64) error {
	_, err := q.db.Exec(ctx, deleteUserRefreshTokens, userID)
	return err
}

const getLinkByAlias = `]
])

edit(base + 'repository/repository.go', [
  [`func (r *UserRepository) RevokeRefreshTokenByHash(ctx context.Context, tokenHash string) error {
	return r.q.RevokeRefreshTokenByHash(ctx, tokenHash)
}`, `// DeleteRefreshToken çıkışta token'ı siler. İptal işareti yalnızca rotasyon içindir; silinen token
// sonradan gelirse "bilinmeyen token" sayılır ve tekrar kullanım (çalıntı) alarmı tetiklenmez.
func (r *UserRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	return r.q.DeleteRefreshTokenByHash(ctx, tokenHash)
}`],
  [`func (r *UserRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID int64) error {
	return r.q.RevokeAllUserRefreshTokens(ctx, userID)
}`, `// DeleteAllRefreshTokens kullanıcının bütün oturum yenileme token'larını siler (parola değişimi, askı,
// tekrar kullanım tespiti). Eski cihazlardan gelen token'lar sonradan yeni oturumu düşürmez.
func (r *UserRepository) DeleteAllRefreshTokens(ctx context.Context, userID int64) error {
	return r.q.DeleteUserRefreshTokens(ctx, userID)
}`]
])

edit(base + 'service/service.go', [
  ['_ = s.users.RevokeAllUserRefreshTokens(ctx, ', '_ = s.users.DeleteAllRefreshTokens(ctx, '],
  ['_ = s.users.RevokeRefreshTokenByHash(ctx, auth.HashRefreshToken(rawRefreshToken))', '_ = s.users.DeleteRefreshToken(ctx, auth.HashRefreshToken(rawRefreshToken))']
])
console.log(miss ? 'misses ' + miss : 'fix1 ok')
