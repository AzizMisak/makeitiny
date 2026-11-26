// Geçici uçtan uca test: yalnızca geçici konteynerlere (pg 55432, redis 56379) bağlı api:18080 ve router:18081.
const { execSync } = require('child_process')
const fs = require('fs')
const API = 'http://127.0.0.1:18080'
const R = 'http://127.0.0.1:18081'
const FRONT = 'http://127.0.0.1:3100'
const sleep = (ms) => new Promise((r) => setTimeout(r, ms))
const results = []
const check = (group, name, ok, extra = '') => {
  results.push({ group, name, ok })
  console.log(`${ok ? 'GEÇTİ' : 'KALDI'}  [${group}] ${name}${ok ? '' : '  ->  ' + String(extra).slice(0, 300)}`)
}
const psql = (sql) => execSync(`docker exec mit-test-pg psql -U makeitiny -d makeitiny -tAc "${sql}"`).toString().trim()
const redis = (cmd) => execSync(`docker exec mit-test-redis redis-cli ${cmd}`).toString().trim()

async function req(method, url, { token, body, cookie, headers: extra } = {}) {
  const headers = { 'content-type': 'application/json', ...(extra || {}) }
  if (token) headers.authorization = `Bearer ${token}`
  if (cookie) headers.cookie = cookie
  const res = await fetch(url, { method, headers, body: body !== undefined ? JSON.stringify(body) : undefined, redirect: 'manual' })
  const buf = Buffer.from(await res.arrayBuffer())
  const text = buf.toString('utf8')
  let json = null
  try { json = JSON.parse(text) } catch {}
  return { status: res.status, json, text, buf, headers: res.headers, loc: res.headers.get('location') || '' }
}
const waitUp = async (url) => {
  for (let i = 0; i < 120; i++) { try { if ((await fetch(url)).ok) return } catch {} await sleep(1000) }
  throw new Error('ayağa kalkmadı ' + url)
}
const cookieOf = (r) => (r.headers.get('set-cookie') || '').split(';')[0]
async function login(email, password) {
  const r = await req('POST', `${API}/auth/login`, { body: { email, password } })
  return { token: r.json?.token, user: r.json?.user, cookie: cookieOf(r), raw: r }
}
const reachTarget = 'https://example.com'

;(async () => {
  await waitUp(`${API}/health`)
  await waitUp(`${R}/health`)

  // ---------------- AUTH ----------------
  let r = await req('POST', `${API}/auth/register`, { body: { email: 'kisi@deneme.test', password: 'deneme1234' } })
  check('auth', 'kayıt, ad e-postadan', r.status === 201 || (r.status === 200 && r.json?.user?.name === 'kisi'), `${r.status} ${r.text}`)
  check('auth', 'kayıt yanıtında ad', r.json?.user?.name === 'kisi', r.text)
  check('auth', 'refresh çerezi HttpOnly', /HttpOnly/i.test(r.headers.get('set-cookie') || ''), r.headers.get('set-cookie'))
  r = await req('POST', `${API}/auth/register`, { body: { email: 'kisi@deneme.test', password: 'deneme1234' } })
  check('auth', 'aynı e-posta ikinci kez kaydolamaz', r.status === 409, `${r.status} ${r.text}`)
  r = await req('POST', `${API}/auth/register`, { body: { email: 'kisa@deneme.test', password: 'abc' } })
  check('auth', 'kısa parola reddi', r.status === 400, `${r.status} ${r.text}`)
  r = await req('POST', `${API}/auth/register`, { body: { email: 'gecersiz', password: 'deneme1234' } })
  check('auth', 'geçersiz e-posta reddi', r.status === 400 && r.json?.code === 'invalid_email', `${r.status} ${r.text}`)
  await req('POST', `${API}/auth/register`, { body: { email: 'yonetici@deneme.test', password: 'deneme1234' } })
  await req('POST', `${API}/auth/register`, { body: { email: 'diger@deneme.test', password: 'deneme1234' } })

  r = await req('POST', `${API}/auth/login`, { body: { email: 'kisi@deneme.test', password: 'yanlis' } })
  check('auth', 'yanlış parola 401', r.status === 401, `${r.status}`)
  r = await req('POST', `${API}/auth/login`, { body: { email: 'yok@deneme.test', password: 'yanlis' } })
  check('auth', 'olmayan e-posta da aynı hata', r.status === 401, `${r.status}`)

  let owner = await login('admin@admin.com', 'admin')
  check('auth', 'tohum sahip hesabı giriş yapar', !!owner.token && owner.user.roles.includes('owner'), owner.raw.text)
  let user = await login('kisi@deneme.test', 'deneme1234')
  const other = await login('diger@deneme.test', 'deneme1234')
  let adminU = await login('yonetici@deneme.test', 'deneme1234')
  r = await req('PUT', `${API}/admin/user-role/${adminU.user.id}`, { token: owner.token, body: { role: 'admin' } })
  check('admin', 'sahip yönetici atar', r.status === 200 && r.json.roles.includes('admin'), r.text)
  adminU = await login('yonetici@deneme.test', 'deneme1234')

  // Refresh rotasyonu ve tekrar kullanım
  const s1 = await login('diger@deneme.test', 'deneme1234')
  const rot = await req('POST', `${API}/auth/refresh`, { cookie: s1.cookie })
  check('auth', 'refresh yeni token verir', rot.status === 200 && !!rot.json?.token, `${rot.status} ${rot.text}`)
  const reuse = await req('POST', `${API}/auth/refresh`, { cookie: s1.cookie })
  check('auth', 'eski refresh tekrar kullanılamaz', reuse.status === 401, `${reuse.status}`)
  const afterReuse = await req('POST', `${API}/auth/refresh`, { cookie: cookieOf(rot) })
  check('auth', 'tekrar kullanım bütün oturumları kapatır', afterReuse.status === 401, `${afterReuse.status}`)

  // Ad
  r = await req('PATCH', `${API}/auth/user`, { token: user.token, body: { name: '  Hasan   Mustafa Misak ' } })
  check('auth', 'ad boşluk düzeltilip kaydedilir', r.status === 200 && r.json.name === 'Hasan Mustafa Misak', r.text)
  r = await req('PATCH', `${API}/auth/user`, { token: user.token, body: { name: 'a' } })
  check('auth', 'tek harfli ad reddi', r.status === 400, `${r.status}`)

  // Parola değişimi
  const pwUser = await login('diger@deneme.test', 'deneme1234')
  r = await req('POST', `${API}/auth/resetpass`, { token: pwUser.token, body: { email: 'diger@deneme.test', newPassword: 'yeni12345' } })
  check('auth', 'kendi parolası için mevcut parola gerekir', r.status === 400, `${r.status} ${r.text}`)
  r = await req('POST', `${API}/auth/resetpass`, { token: pwUser.token, body: { email: 'diger@deneme.test', currentPassword: 'yanlis', newPassword: 'yeni12345' } })
  check('auth', 'yanlış mevcut parola reddi', r.status === 400 && r.json?.code === 'invalid_current_password', `${r.status} ${r.text}`)
  r = await req('POST', `${API}/auth/resetpass`, { token: pwUser.token, body: { email: 'diger@deneme.test', currentPassword: 'deneme1234', newPassword: 'yeni12345' } })
  check('auth', 'parola değişir', r.status === 200, `${r.status} ${r.text}`)
  r = await req('GET', `${API}/auth/user`, { token: pwUser.token })
  check('auth', 'parola değişince eski erişim token geçersiz', r.status === 401, `${r.status}`)
  const otherNew = await login('diger@deneme.test', 'yeni12345')
  check('auth', 'yeni parolayla giriş', !!otherNew.token, otherNew.raw.text)

  // Çıkış
  const lo = await login('diger@deneme.test', 'yeni12345')
  r = await req('POST', `${API}/auth/logout`, { token: lo.token, cookie: lo.cookie })
  check('auth', 'çıkış 200', r.status === 200, `${r.status}`)
  r = await req('POST', `${API}/auth/refresh`, { cookie: lo.cookie })
  check('auth', 'çıkıştan sonra refresh çalışmaz', r.status === 401, `${r.status}`)
  r = await req('GET', `${API}/auth/user`, { token: lo.token })
  check('auth', 'çıkıştan sonra erişim token çalışmaz', r.status === 401, `${r.status}`)

  // ---------------- LINKS ----------------
  r = await req('POST', `${API}/links`, { token: user.token, body: { target: 'example.com/a' } })
  const L1 = r.json
  check('links', 'oluşturma, https eklenir, 6 karakter kod', r.status === 201 || r.status === 200 ? L1.target === 'https://example.com/a' && /^[0-9A-Za-z]{6}$/.test(L1.alias) : false, `${r.status} ${r.text}`)
  r = await req('POST', `${API}/links`, { token: user.token, body: { target: 'https://example.com', alias: 'benim-adim' } })
  check('links', 'ücretsiz pakette özel kısa ad reddi', r.status === 403, `${r.status} ${r.text}`)
  r = await req('POST', `${API}/links`, { token: owner.token, body: { target: 'https://example.com', alias: 'sahip-adi' } })
  check('links', 'sahip özel kısa ad alır', (r.status === 201 || r.status === 200) && r.json.alias === 'sahip-adi', `${r.status} ${r.text}`)
  r = await req('POST', `${API}/links`, { token: owner.token, body: { target: 'https://example.com', alias: 'sahip-adi' } })
  check('links', 'alınmış kısa ad 409', r.status === 409, `${r.status}`)
  r = await req('GET', `${API}/links/check-alias/sahip-adi`, { token: user.token })
  check('links', 'check-alias alınmış', r.json?.available === false, r.text)

  for (const t of ['http://127.0.0.1', 'http://127.1', 'http://0x7f.0.0.1', 'http://2130706433', 'http://[::1]', 'http://169.254.169.254', 'http://api:8080', 'http://redis:6379', 'http://localhost', `${FRONT}/r/abc`]) {
    r = await req('POST', `${API}/links`, { token: user.token, body: { target: t } })
    check('ssrf', `reddedilir: ${t}`, r.status === 400, `${r.status} ${r.text}`)
  }
  r = await req('POST', `${API}/links`, { token: user.token, body: { target: 'https://example.com', password: '   ' } })
  check('links', 'boşluk parola reddi', r.status === 400, `${r.status}`)
  r = await req('POST', `${API}/links`, { token: user.token, body: { target: 'https://example.com', waitEnabled: true, waitSeconds: 100000 } })
  check('links', 'aşırı bekleme süresi reddi', r.status === 400, `${r.status} ${r.text}`)

  // IDOR
  r = await req('GET', `${API}/links/${L1.id}`, { token: otherNew.token })
  check('idor', 'başkasının linki okunamaz', r.status === 404, `${r.status}`)
  r = await req('PATCH', `${API}/links/${L1.id}`, { token: otherNew.token, body: { target: 'https://evil.example' } })
  check('idor', 'başkasının linki değiştirilemez', r.status === 404, `${r.status}`)
  r = await req('DELETE', `${API}/links/${L1.id}`, { token: otherNew.token })
  check('idor', 'başkasının linki silinemez', r.status === 404, `${r.status}`)
  r = await req('GET', `${API}/links/logs/${L1.id}`, { token: otherNew.token })
  check('idor', 'başkasının ziyaretleri okunamaz', r.status === 404, `${r.status}`)
  r = await req('GET', `${API}/links/logs/${L1.id}/export`, { token: otherNew.token })
  check('idor', 'başkasının ziyaretleri indirilemez', r.status === 404, `${r.status}`)
  r = await req('GET', `${API}/admin/users`, { token: user.token })
  check('idor', 'kullanıcı yönetim uç noktasına giremez', r.status === 403, `${r.status}`)

  // Kota yarışı
  await req('PATCH', `${API}/admin/plans/1`, { token: owner.token, body: { maxLinks: 3 } })
  const quotaUser = await login('kisi@deneme.test', 'deneme1234')
  const have = Number(psql("SELECT COUNT(*) FROM links l JOIN users u ON u.id=l.user_id WHERE u.email='kisi@deneme.test'"))
  const burst = await Promise.all(Array.from({ length: 8 }, () => req('POST', `${API}/links`, { token: quotaUser.token, body: { target: 'https://example.com/kota' } })))
  const made = burst.filter((x) => x.status === 201 || x.status === 200).length
  check('links', `kota yarışında sınır aşılmaz (vardı ${have}, sınır 3)`, have + made === 3, `oluşan ${made}, durumlar ${burst.map((x) => x.status)}`)
  await req('PATCH', `${API}/admin/plans/1`, { token: owner.token, body: { maxLinks: 50 } })

  // ---------------- ROUTER ----------------
  const mk = async (body, token = owner.token) => (await req('POST', `${API}/links`, { token, body: { target: reachTarget, ...body } })).json
  const direct = await mk({})
  r = await req('GET', `${R}/${direct.alias}`)
  const online = r.status === 302 && r.loc === reachTarget
  check('router', 'doğrudan yönlendirme', online, `${r.status} ${r.loc}`)
  if (!online) console.log('   (internet yoksa hedef denetimine bağlı denemeler başarısız görünebilir)')
  r = await req('GET', `${R}/${direct.alias}/`)
  check('router', 'sonda eğik çizgi de çalışır', r.loc === reachTarget, `${r.status} ${r.loc}`)
  r = await req('GET', `${R}/${direct.alias}/bilinmeyen/yol`)
  check('router', 'bilinmeyen alt yol hata ekranına', r.status === 302 && r.loc.includes('/error'), `${r.status} ${r.loc}`)
  r = await req('GET', `${R}/`)
  check('router', '/r/ kökü hata ekranına', r.status === 302 && r.loc.includes('/error'), `${r.status} ${r.loc}`)
  r = await req('GET', `${R}/olmayan-kod`)
  check('router', 'olmayan kod not_found', r.loc.endsWith('reason=not_found'), r.loc)
  r = await req('HEAD', `${R}/${direct.alias}`)
  check('router', 'HEAD isteği cevaplanır', r.status === 302 || r.status === 200, `${r.status}`)
  r = await req('GET', `${R}/${direct.alias}/preview`)
  check('router', 'önizleme hedef içermez', r.status === 200 && !JSON.stringify(r.json).includes('example.com'), r.text)

  // Tıklama sınırı yarışı
  const lim = await mk({ maxClicks: 3 })
  const hits = await Promise.all(Array.from({ length: 15 }, () => req('GET', `${R}/${lim.alias}`)))
  const okHits = hits.filter((h) => h.loc === reachTarget).length
  const limHits = hits.filter((h) => h.loc.endsWith('reason=limit_reached')).length
  check('router', `15 eşzamanlı tıklamada sınır 3 aşılmaz`, okHits === 3 && limHits === 12, `hedefe ${okHits}, sınır ${limHits}`)

  // Bekleme bileti + sınır 1
  const w = await mk({ waitEnabled: true, waitSeconds: 2, maxClicks: 1 })
  r = await req('GET', `${R}/${w.alias}`)
  const wt = r.loc.includes('/w/') ? new URL(r.loc).searchParams.get('t') : ''
  check('wait', 'bilet ile bekleme sayfası', !!wt, r.loc)
  if (wt) {
    r = await req('GET', `${R}/${w.alias}/wait?t=${wt}`)
    check('wait', 'bilgi döner (sınır dolsa bile)', r.status === 200 && r.json.remainingSeconds >= 1, `${r.status} ${r.text}`)
    r = await req('POST', `${R}/${w.alias}/continue`, { body: { ticket: wt } })
    check('wait', 'erken devam 425', r.status === 425, `${r.status}`)
    r = await req('POST', `${R}/${w.alias}/continue`, { body: { ticket: 'uydurma' } })
    check('wait', 'uydurma bilet 410', r.status === 410, `${r.status} ${r.text}`)
    await sleep(2300)
    const [c1, c2] = await Promise.all([
      req('POST', `${R}/${w.alias}/continue`, { body: { ticket: wt } }),
      req('POST', `${R}/${w.alias}/continue`, { body: { ticket: wt } })
    ])
    check('wait', 'eşzamanlı iki devamdan yalnızca biri hedef alır', [c1, c2].filter((c) => c.status === 200).length === 1, `${c1.status} ${c2.status}`)
    r = await req('GET', `${R}/${w.alias}`)
    check('wait', 'ikinci ziyaretçi sınır ekranı', r.loc.endsWith('reason=limit_reached'), r.loc)
  }
  r = await req('GET', `${R}/${w.alias}/wait`)
  check('wait', 'biletsiz bekleme isteği 400', r.status === 400, `${r.status} ${r.text}`)

  // Parolalı link
  const lk = await mk({ password: 'gizli123' })
  r = await req('GET', `${R}/${lk.alias}`)
  check('unlock', 'parolalı link /unlock ekranına', r.loc.includes('/unlock/'), r.loc)
  r = await req('POST', `${R}/${lk.alias}/unlock`, { body: { password: 'yanlis' } })
  check('unlock', 'yanlış parola 401', r.status === 401, `${r.status}`)
  r = await req('POST', `${R}/${lk.alias}/unlock`, { body: { password: 'gizli123' } })
  check('unlock', 'doğru parola hedefi verir', r.status === 200 && r.json.target === reachTarget, `${r.status} ${r.text}`)
  let got429 = false
  for (let i = 0; i < 12; i++) {
    const x = await req('POST', `${R}/${lk.alias}/unlock`, { body: { password: 'yanlis' } })
    if (x.status === 429) { got429 = true; break }
  }
  check('unlock', 'parola denemesi sınırlanır (429)', got429)
  r = await req('POST', `${R}/${direct.alias}/unlock`, { body: { password: 'x' } })
  check('unlock', 'parolasız linkte unlock 400', r.status === 400, `${r.status} ${r.text}`)
  r = await req('POST', `${R}/${lk.alias}/unlock`, { body: 'x'.repeat(70000) })
  check('unlock', 'büyük gövde reddi', r.status === 400 || r.status === 413, `${r.status}`)

  // Durumlar
  const blk = await mk({})
  await req('PATCH', `${API}/admin/links/${blk.id}`, { token: adminU.token, body: { securityStatus: 'malicious' } })
  r = await req('GET', `${R}/${blk.alias}`)
  check('router', 'zararlı link engelli ekranı', r.loc.endsWith('reason=blocked'), r.loc)
  r = await req('PATCH', `${API}/links/${blk.id}`, { token: owner.token, body: { isActive: true } })
  check('router', 'engelli linki sahibi yayına alamaz', r.status === 403 || (r.json && r.json.isActive === false), `${r.status} ${r.text}`)
  const off = await mk({})
  await req('PATCH', `${API}/links/${off.id}`, { token: owner.token, body: { isActive: false } })
  r = await req('GET', `${R}/${off.alias}`)
  check('router', 'kapalı link devre dışı ekranı', r.loc.endsWith('reason=disabled'), r.loc)
  const exp = await mk({})
  psql(`UPDATE links SET expires_at = NOW() - INTERVAL '1 minute' WHERE public_id='${exp.id}'`)
  redis(`DEL link:${exp.alias}`)
  r = await req('GET', `${R}/${exp.alias}`)
  check('router', 'süresi dolan link süre ekranı', r.loc.endsWith('reason=expired'), r.loc)

  // Hedef değişimi ve önbellek
  const ch = await mk({})
  await req('GET', `${R}/${ch.alias}`)
  await req('PATCH', `${API}/links/${ch.id}`, { token: owner.token, body: { target: 'https://example.org' } })
  r = await req('GET', `${R}/${ch.alias}`)
  check('router', 'hedef değişince yeni hedefe gider (önbellek temizlenir)', r.loc === 'https://example.org' || r.loc.includes('unreachable'), r.loc)

  // Anonim
  r = await req('POST', `${API}/links/anonymous`, { body: { target: 'https://example.com/anon', ttlHours: 1 }, headers: { 'x-real-ip': '198.51.100.7' } })
  check('anon', 'anonim link', r.status === 201 || r.status === 200, `${r.status} ${r.text}`)
  r = await req('POST', `${API}/links/anonymous`, { body: { target: 'https://example.com/anon2' }, headers: { 'x-real-ip': '198.51.100.7' } })
  check('anon', 'aynı IP ikinci anonim link 429', r.status === 429, `${r.status} ${r.text}`)

  // ---------------- ADMIN ----------------
  const ownerId = owner.user.id
  r = await req('PATCH', `${API}/admin/users/${ownerId}`, { token: adminU.token, body: { isActive: false } })
  check('admin', 'yönetici sahibi askıya alamaz', r.status === 403, `${r.status}`)
  r = await req('PUT', `${API}/admin/user-role/${ownerId}`, { token: adminU.token, body: { role: 'user' } })
  check('admin', 'yönetici rol değiştiremez', r.status === 403, `${r.status}`)
  r = await req('POST', `${API}/auth/resetpass`, { token: adminU.token, body: { email: 'admin@admin.com', newPassword: 'ele-gecir' } })
  check('admin', 'yönetici sahibin parolasını değiştiremez', r.status === 403, `${r.status}`)
  r = await req('PATCH', `${API}/admin/users/${adminU.user.id}`, { token: owner.token, body: { planId: 2 } })
  check('admin', 'yöneticiye paket atanamaz', r.status === 400, `${r.status} ${r.text}`)
  r = await req('PATCH', `${API}/admin/plans/2`, { token: owner.token, body: { priceMonthly: -1 } })
  check('admin', 'geçersiz paket değeri reddi', r.status === 400, `${r.status}`)
  r = await req('PATCH', `${API}/admin/plans/2`, { token: adminU.token, body: { priceMonthly: 50 } })
  check('admin', 'yönetici paket düzenleyemez', r.status === 403, `${r.status}`)
  r = await req('GET', `${API}/admin/links?status=bozuk`, { token: adminU.token })
  check('admin', 'geçersiz durum süzgeci 400', r.status === 400, `${r.status}`)
  r = await req('GET', `${API}/admin/stats`, { token: adminU.token })
  check('admin', 'istatistik', r.status === 200 && r.json.totalUsers >= 4, r.text)
  r = await req('DELETE', `${API}/auth/user`, { token: owner.token })
  check('admin', 'sahip hesabı silinemez', r.status === 403, `${r.status}`)

  // ---------------- DIŞA AKTARMA ----------------
  const exUser = await login('kisi@deneme.test', 'deneme1234')
  r = await req('GET', `${API}/links/export?format=csv`, { token: exUser.token })
  check('export', 'CSV BOM + başlık + ek', r.status === 200 && r.buf[0] === 0xef && r.text.includes('Kısa adres') && (r.headers.get('content-disposition') || '').includes('attachment'), `${r.status}`)
  r = await req('GET', `${API}/links/export?format=json`, { token: exUser.token })
  check('export', 'JSON dizi', Array.isArray(r.json), r.text.slice(0, 100))
  r = await req('GET', `${API}/links/export?format=xml`, { token: exUser.token })
  check('export', 'geçersiz biçim 400', r.status === 400, `${r.status}`)

  // ---------------- QR ----------------
  r = await req('POST', `${API}/links/${L1.id}/qr`, { token: exUser.token })
  check('qr', 'ücretsiz pakette QR reddi', r.status === 403, `${r.status}`)
  r = await req('POST', `${API}/links/${direct.id}/qr`, { token: owner.token })
  check('qr', 'sahip QR üretir', r.status === 200 && r.json.hasQr === true, `${r.status} ${r.text}`)
  r = await req('GET', `${API}/links/${direct.id}/qr`, { token: owner.token })
  check('qr', 'QR PNG döner', r.status === 200 && r.buf.slice(1, 4).toString() === 'PNG', `${r.status} ${r.headers.get('content-type')}`)
  r = await req('DELETE', `${API}/links/${direct.id}/qr`, { token: owner.token })
  check('qr', 'QR silinir', r.status === 204 || r.status === 200, `${r.status}`)

  // ---------------- BAN + WORKER ----------------
  const banTarget = await login('kisi@deneme.test', 'deneme1234')
  const banLink = (await req('POST', `${API}/links`, { token: banTarget.token, body: { target: reachTarget } })).json
  await req('GET', `${R}/${banLink.alias}`)
  const until = new Date(Date.now() + 4000).toISOString()
  r = await req('PATCH', `${API}/admin/users/${banTarget.user.id}`, { token: adminU.token, body: { isActive: false, banReason: 'deneme', bannedUntil: until } })
  check('ban', 'yönetici kullanıcıyı askıya alır', r.status === 200 && r.json.isActive === false, `${r.status} ${r.text}`)
  r = await req('GET', `${API}/auth/user`, { token: banTarget.token })
  check('ban', 'askıdaki kullanıcı 403 account_suspended', r.status === 403 && r.json.code === 'account_suspended', `${r.status} ${r.text}`)
  r = await req('GET', `${R}/${banLink.alias}`)
  check('ban', 'askıdaki kullanıcının linki çalışmaz', r.loc.includes('/error'), r.loc)

  // Takma ad değişimi aktarım öncesi
  const mv = await mk({})
  await req('GET', `${R}/${mv.alias}`)
  await req('GET', `${R}/${mv.alias}`)
  await req('PATCH', `${API}/links/${mv.id}`, { token: owner.token, body: { alias: 'yeni-ad-' + Date.now().toString(36) } })

  console.log('   worker turu bekleniyor (en fazla 40 sn)...')
  let unbanned = false
  for (let i = 0; i < 40 && !unbanned; i++) {
    await sleep(1000)
    unbanned = psql(`SELECT is_active FROM users WHERE public_id='${banTarget.user.id}'`) === 't'
  }
  await sleep(1500)
  check('ban', 'süresi dolan askı kendiliğinden kalkar', unbanned)
  r = await req('GET', `${R}/${banLink.alias}`)
  check('ban', 'askı kalkınca link yeniden çalışır', r.loc === reachTarget, r.loc)
  const mvClicks = psql(`SELECT clicks FROM links WHERE public_id='${mv.id}'`)
  check('worker', 'takma ad değişse de tıklamalar doğru linke yazılır', mvClicks === '2', `clicks=${mvClicks}`)
  const limClicks = psql(`SELECT clicks FROM links WHERE public_id='${lim.id}'`)
  check('worker', 'sınırlı link tıklaması 3 yazıldı', limClicks === '3', `clicks=${limClicks}`)
  const logCount = Number(psql(`SELECT COUNT(*) FROM logs WHERE link_alias='${lim.alias}'`))
  check('worker', 'ziyaret kayıtları yazıldı', logCount === 3, `logs=${logCount}`)
  check('worker', 'kuyrukta işlenmemiş kayıt kalmadı', redis('LLEN logs:queue') === '0', redis('LLEN logs:queue'))
  check('worker', 'askıda tıklama kopyası kalmadı', redis('KEYS clicks:deltas:proc:*') === '', redis('KEYS clicks:deltas:proc:*'))

  // Zaman damgası biçimi
  r = await req('GET', `${API}/links/logs/${lim.id}`, { token: owner.token })
  check('api', 'ziyaret zamanı UTC RFC3339', /Z$/.test(r.json?.items?.[0]?.clickedAt || ''), r.json?.items?.[0]?.clickedAt)

  // Hesap silme
  const del = await login('diger@deneme.test', 'yeni12345')
  await req('POST', `${API}/links`, { token: del.token, body: { target: 'https://example.com/silinecek' } })
  r = await req('DELETE', `${API}/auth/user`, { token: del.token })
  check('auth', 'hesap silinir', r.status === 204 || r.status === 200, `${r.status}`)
  check('auth', 'silinen hesabın linkleri gider', psql("SELECT COUNT(*) FROM links WHERE target_url='https://example.com/silinecek'") === '0')

  const failed = results.filter((x) => !x.ok)
  console.log(`\n${results.length - failed.length}/${results.length} geçti`)
  fs.writeFileSync('.tmp-e2e-result.json', JSON.stringify(failed, null, 1))
  process.exit(failed.length ? 1 : 0)
})().catch((e) => { console.error(e); process.exit(2) })
