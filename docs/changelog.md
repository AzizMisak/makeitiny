# Geliştirme Günlüğü (Yapılan İşler)

Bu doküman, MakeItTiny projesinde gerçekleştirilen altyapı, mimari ve geliştirme adımlarını özet olarak takip etmek amacıyla tutulmaktadır.

---

## 1. Port Yapılandırması ve Path Bazlı Nginx Yönlendirmesi
- **Geliştirme Portları:** `docker-compose.yml` üzerinde geliştirme sürecinde doğrudan test yapılabilmesi adına tüm servislerin portları (`80`, `3000`, `8080`, `8081`, `8082`, `5432`, `6379`) dışarıya açıldı.
- **Nginx Güvenliği ve Yönlendirme (`nginx/nginx.conf`):**
  - Tanımsız `Host` başlıkları ve dış adreslerden gelen istekleri engellemek için `default_server` bloğu (`return 444`) eklendi; yalnızca `localhost` ve `127.0.0.1` isteklerine izin verildi.
  - Subdomain karmaşası kaldırılarak Docker iç ağı üzerinden 3 ana path yönlendirmesi kuruldu:
    - `/api/` → `http://api:8080/`
    - `/r/` → `http://router:8081/`
    - `/` → `http://frontend:3000`
- **Nuxt ve Swagger Uyumlaması:**
  - `Frontend/nuxt.config.ts` içindeki `apiBase` (`http://localhost/api`) ve `routerBase` (`http://localhost/r`) çakışmaları giderildi.
  - `openapi/api.yaml` ve `openapi/router.yaml` sunucu listelerine hem Nginx proxy hem de doğrudan port adresleri eklendi; `router` servisinin CORS izin listesi genişletildi.

---

## 2. OpenAPI Sözleşmelerinin ve Şemaların Tamamlanması
- **Endpoint Kontrolü (`openapi/api.yaml` & `openapi/router.yaml`):**
  - Kullanıcının kayıt olma, giriş/çıkış yapma, şifre sıfırlama, **hesap silme (`DELETE /auth/user`)** ve kısa link oluşturma, listeleme, görüntüleme, düzenleme ve silme işlemlerini eksiksiz yapabilmesi için uç noktalar tamamlandı.
  - JWT kimlik doğrulama şeması (`bearerAuth`) ve tüm istek/yanıt veri modelleri (`User`, `AuthResponse`, `LoginRequest`, `RegisterRequest`, `ResetPasswordRequest`, `LinkItem`, `CreateLinkRequest`, `UpdateLinkRequest`, `AliasCheckResult`, `UnlockLinkRequest`, `UnlockLinkResponse`, `LinkPreview`, `ApiError`) tanımlandı.
  - `oapi-codegen` ile `Backend/api` ve `Backend/router` servislerinin sunucu arayüzleri ve Go veri modelleri yeniden üretildi.

---

## 3. Veritabanı Altyapısı (`goose` + `sqlc`)
- **Goose Migration (`Backend/db/migrations/00001_init.sql`):**
  - Ortak `Backend/db/migrations` dizini oluşturuldu ve başlangıç tabloları (`users` ve `links`) indeksleriyle birlikte (`idx_links_alias`, `idx_links_user_id`) PostgreSQL üzerinde ayağa kaldırıldı.
- **sqlc Kurulumu (`Backend/api` & `Backend/router`):**
  - Her iki servisin sadece kendi sorumluluk alanına giren SQL sorgularını barındırması için `Backend/api/sqlc.yaml` ve `Backend/router/sqlc.yaml` ayrı ayrı yapılandırıldı.
  - `pgx/v5` tabanlı tip güvenli veritabanı erişim kodları (`internal/db`) her iki servis için üretildi.

---

## 4. Uygulama İçi Mimari Kararların Netleştirilmesi (ADR-010)
- **Auth & Varsayılan İsim:** JWT Bearer Token kullanımına ve kayıt anında isim belirtilmezse e-postanın `@` öncesi kısmının varsayılan kullanıcı adı olarak atanmasına karar verildi.
- **Şifreleme & Kısa Kod:** Parolalar (kullanıcı ve şifreli link) için `Argon2id`, otomatik kısa link üretimi için rastgele `Base62` seçildi.
- **Redis Hash & Senkronizasyon:** Redis'te link verilerinin `link:{alias}` anahtarıyla **Redis Hash (`HSET`/`HGETALL`/`HINCRBY`)** olarak tutulması ve işlemlerde önce PostgreSQL ardından Redis'in güncellenmesi kararlaştırıldı.
- **Şifreli Link Ekranı:** Şifreli linklere (`GET /r/{link}`) tıklandığında kullanıcının Frontend'deki `/unlock/{link}` sayfasına yönlendirilmesi kararlaştırıldı.

---

## 5. `api` Servisinin Tamamlanması (Aşama 1 & Aşama 2)
- **Altyapı Bağlantıları (`Backend/api/main.go`):**
  - PostgreSQL (`pgxpool`), başlangıçta otomatik çalışan `goose` migration ve Redis (`go-redis/v9`) bağlantıları eklendi.
- **Güvenlik ve Kimlik Doğrulama (`Backend/api/internal/auth`):**
  - `password.go`: `Argon2id` PHC formatında parola hash'leme (`HashPassword`) ve sabit zamanlı doğrulama (`VerifyPassword`) fonksiyonları yazıldı.
  - `jwt.go`: `HS256` tabanlı JWT üretimi (`GenerateToken`), doğrulaması (`ParseToken`) ve `Authorization: Bearer` / çerez okuma (`ExtractToken`) yardımcıları eklendi.
- **Repository, Service ve Handler Katmanları (`Backend/api/internal/{repository,service,handler}`):**
  - **Auth Uç Noktaları:** `POST /auth/register` (varsayılan isim e-postanın `@` öncesi), `POST /auth/login`, `POST /auth/logout`, `POST /auth/resetpass`, `GET /auth/user` ve `DELETE /auth/user` (kullanıcı silindiğinde Redis üzerindeki link önbellekleri de temizlenir) tamamlandı.
  - **Link Uç Noktaları & Redis Hash:** Kriptografik rastgele 6 karakterli `Base62` alias üretimi ile `POST /links`, `GET /links`, `GET /links/check-alias/{alias}`, `GET /links/{id}`, `PATCH /links/{id}` ve `DELETE /links/{id}` uç noktaları tamamlandı. Tüm yazma/güncelleme/silme işlemlerinde önce PostgreSQL, hemen ardından Redis Hash (`HSET link:{alias}` / `DEL link:{alias}`) güncellenecek şekilde kurgulandı.

---

## 6. `router` Servisi ve `Frontend` Entegrasyonu (Aşama 3 & Aşama 4)
- **`router` Servisi (`Backend/router`):**
  - PostgreSQL (`pgxpool`) ve Redis (`go-redis/v9`) bağlantıları eklendi.
  - `GET /{link}`: Önce Redis Hash (`HGETALL link:{alias}`), bulunamazsa PostgreSQL üzerinden link çözümlenir. Şifresiz linklerde önce PostgreSQL'de tıklama sayısı artırılır, hemen ardından Redis Hash üzerinde `HINCRBY link:{alias} clicks 1` çalıştırılır ve `302 Found` ile hedef adrese yönlendirilir. Şifreli linklerde kullanıcı `302 Found` ile Frontend'deki `/unlock/{link}` sayfasına yönlendirilir.
  - `POST /{link}/unlock`: Parolayı `Argon2id` ile doğrular, tıklama sayısını (önce PostgreSQL, ardından Redis) artırır ve hedef URL'i döner.
  - `GET /{link}/preview`: Kısa linkin önizleme bilgisini döner (şifreli linklerde hedef URL gizlenir).
- **`Frontend` Entegrasyonu (`Frontend/app`):**
  - `useApi.ts`, `useAuth.ts` ve `useLinks.ts` içerisindeki sahte (mock) veriler kaldırılarak gerçek Backend uç noktalarına (`JWT` Bearer yetkilendirmesiyle) bağlandı.
  - `LoginForm.vue`, `RegisterForm.vue`, `SettingsForm.vue`, `CreateLinkForm.vue` ve `LinkRow.vue` bileşenleri gerçek API çağrıları ve satır içi link düzenleme (`PATCH /links/{id}`) özelliğiyle güncellendi.
  - Şifreli kısa linklerin kilidini açmak için `Frontend/app/pages/unlock/[alias].vue` sayfası oluşturuldu.

---

## 7. Frontend Kullanıcı Deneyimi (UX) ve Doğrulama Revizyonu
- **Yükleniyor (Loading) Durumu (`useLinks.ts`, `LinkList.vue`, `dashboard/index.vue`, `dashboard/links/index.vue`):**
  - Veriler API'den çekilirken erken "Henüz link yok" mesajı gösterilmesi engellendi; dönen yükleme ikonu ile `"Linkler yükleniyor..."` göstergesi eklendi.
- **Link Detay Sayfası ve Şifre Kaldırma (`pages/dashboard/links/[id].vue`, `LinkRow.vue`):**
  - Her kısa link için bağımsız detay ve yönetim sayfası (`/dashboard/links/[id]`) oluşturuldu; listede linkin üzerine tıklandığında bu sayfaya yönlendirilmesi sağlandı.
  - Şifreli linklerin parolasını tek tıkla kaldırma (**"Şifreyi Kaldır"** -> `PATCH /links/{id}` `password: ""`) ve parola değiştirme/ekleme akışı eklendi.
- **Varsayılan `https://` Tamamlama (`useLinks.ts`, `CreateLinkForm.vue`, `Backend/api/internal/service/service.go`):**
  - Link oluşturma ve düzenleme formlarında `https://` yazma zorunluluğu kaldırıldı; protokol belirtilmezse hem Frontend hem Backend tarafında otomatik olarak `https://` eklenecek şekilde güncellendi.
- **Parola Doğrulama ve Çift Giriş Kontrolü (`SettingsForm.vue`, `RegisterForm.vue`, `service.go`, `handler.go`):**
  - Kayıt ve şifre değiştirme ekranlarındaki minimum parola uzunluğu metinleri (`En az 4 karakter`) Backend kuralıyla eşitlendi.
  - Hesap ayarlarında şifre değiştirirken parolanın iki kez girilmesi (`Yeni Parola` ve `Yeni Parola (Tekrar)`) zorunlu kılındı.
  - Minimum karakterin altında şifre girildiğinde dönen yanıltıcı hata mesajı düzeltilerek `password_too_short` (`"parola en az 4 karakter olmalıdır"`) hatası ayrıştırıldı.
- **Ana Sayfa Oturum Kontrolleri (`LandingHero.vue`, `LandingCta.vue`, `auth/login.vue`, `auth/register.vue`):**
  - Kullanıcı giriş yapmış durumdayken ana sayfadaki "Giriş yap" ve "Ücretsiz hesap aç / başla" butonları gizlenerek yerine **"Panele git"** ve **"Yeni link kısalt"** butonları eklendi.

---

## 8. Yapay Zeka Asistanı Rehberi (`agents.md`)
- Kök dizine projenin amacını, mimari yapısını, yönlendirme/port kurallarını, teknik standartlarını (`sqlc`, `goose`, `Redis Hash`, `Argon2id`, `Base62`) ve kesin çalışma/iletişim kurallarını (istenmeyen iş yapmama, belirsizlikleri tartışma, commit atmama, tur sonu özeti verme) içeren `agents.md` dosyası eklendi.

---

## 9. Güvenlik, Performans (Write-Behind), Anonim Kısaltma, QR Depolama ve Frontend UX Dönüşümü
- **Veritabanı (`00004_user_public_id_bans_anon_qr.sql`) & `sqlc`:**
  - `users` tablosuna 16 karakterli `public_id`, `ban_reason`, `banned_at`, `banned_until` alanları ve `owner` silinmesini engelleyen `trg_prevent_owner_delete` trigger'ı eklendi.
  - `links` tablosuna `is_anonymous` ve `created_by_ip` eklendi; `logs.link_id` `ON DELETE SET NULL` yapılarak `link_alias`, `link_public_id`, `event_type` ile denetim izi korundu. Refresh token rotasyonu için `refresh_tokens` tablosu oluşturuldu.
- **Kimlik Doğrulama & Ban (`Backend/api`):**
  - 15 dakikalık kısa ömürlü JWT (`JTI` destekli) + `HttpOnly` Refresh Token rotasyonu (`POST /api/auth/refresh`, replay tespitli oturum iptali) devreye alındı.
  - Admin/Owner kullanıcı banlama (`reason` + süreli/kalıcı) ve Redis üzerinden anlık JWT/oturum geçersiz kılma (`user:revoked_before`, `jwt:blacklist`) eklendi.
- **Redis Write-Behind & Router Optimizasyonu (`Backend/api`, `Backend/router`):**
  - Router sıcak yönlendirme yolu PostgreSQL'e senkron yazma yerine Redis Pipeline (`HINCRBY`, `RPUSH logs:queue`) + `clicks:deltas` modeline geçirildi; `api` servisine 30 saniyelik toplu PostgreSQL aktarım worker'ı (`worker.go`) eklendi.
  - Şifreli linklerin yalnızca `/unlock` başarısında sayım yapması ve hedef URL erişilebilirlik durumunun 5 dakika Redis'te önbelleklenmesi sağlandı.
- **Anonim Kısaltma & QR Kod Sistemi:**
  - `POST /api/links/anonymous` ile IP bazlı günde 1 link (`429` limiti, en fazla 48 saat ömür, parola/QR/bekleme desteği) eklendi.
  - `go-qrcode` ile paylaşımlı `./Backend/storage/qrcodes` disk hacmine PNG üretimi (`POST/GET/DELETE /api/links/{id}/qr`), alias değişiminde otomatik yenileme ve silmede temizlik kurgulandı.
- **Frontend & Depo Yapılandırması:**
  - `.gitattributes` ile GitHub Linguist dil algılaması Go odaklı yapılandırıldı.
  - Frontend'de zero-FOUC gece modu, Kişisel / Yönetim / Owner görsel çalışma alanı ayrımı, otomatik 401 refresh & retry, link silme onay modalı, anonim kısaltma bölümü (kopyalama uyarılı), QR yönetimi ve admin ban/moderasyon ekranları tamamlandı.

---

## 10. Arayüzün Baştan Tasarlanması (Panel, Landing ve Ziyaretçi Sayfaları)
- **Tasarım sistemi (ADR-013):** Kesim matı teması, yeni renk token'ları, yeni yazı tipleri ve `components/ui` bileşenleri (anahtar, seçim şeridi, sekme, pencere, menü, iskelet yükleme, sayfalama) hazırlandı. Checkbox'lar kaldırıldı.
- **Kullanıcı paneli:** Her sayfa kendi düzenini aldı. Link detayı Analiz, Ziyaretler, QR kodu ve Ayarlar sekmelerine ayrıldı. Tek "Yeni link" ve tek tema düğmesi kaldı, yan menüdeki paket/kota kutusu kaldırıldı.
- **Yönetim paneli:** Ayrı iskelete taşındı. "Hızlı moderasyon" yerine beş seçenekli inceleme kararı geldi. Owner rol sayfası tüm yönetim ekibini listeler.
- **Landing:** Çalışan kısaltma tezgâhı, denenebilir ziyaretçi ekranları, tıklama kaydı gösterimi ve şerit metre gezgini eklendi.
- **Doğruluk düzeltmeleri:** Grafikler en yeni 500 ziyaretten hesaplanıyor, parola değişince oturumların kapanacağı söyleniyor, parolalı linklerle ilgili eski tıklama notu ve ücretsiz pakette olmayan özelliklerin tanıtımı kaldırıldı.

---

## 11. Ad Değiştirme, Paket Muafiyeti ve Analiz Düzeni
- **Ad değiştirme (ADR-014):** `PATCH /auth/user` eklendi (spec, `UpdateUserName` sorgusu, servis, handler). Hesap ayarlarında ad, satırın içinde düzenleniyor.
- **Yetkili hesaplar:** Yönetici ve sahibe paket atanması backend'de `plan_not_applicable` ile engellendi. Genel bakış, Paketim, Hesap ayarları ve yönetim ekranları "Pakete tabi değil" gösteriyor.
- **"Sen" işareti:** Yönetimde kullanıcılar, link sahibi ve roller listelerinde oturumdaki kişi işaretleniyor; kendi kullanıcı sayfasında "Bu sensin" uyarısı çıkıyor.
- **Analiz:** Cihaz, Tarayıcı, İşletim sistemi ve Ülkeler bölümleri ayrı, ince çerçeveli kutulara alındı.

---

## 12. Denetim Sonrası Güvenlik, Yarış Durumu ve Arayüz Tutarlılık Düzeltmeleri
- **Worker & Oturum Dayanıklılığı (`Backend/api`):**
  - `FlushLogsQueue` için NUL bayt temizleme (`sanitizePGText`) ve kalıcı PostgreSQL hatalarında (`22xxx`/`23xxx`) satır bazlı düşüş (`insertLogsOneByOne`) eklendi; tek bozuk log kaydının tüm kuyruğu kilitlemesi önlendi.
  - `RevokeActiveRefreshToken` sorgusuna `expires_at > NOW()` eklendi; süresi dolmuş refresh token ile çakışan isteklerin diğer cihazlardaki aktif oturumları kapatması engellendi.
  - `PendingClicks` ve `TotalPendingClicks` aktarım sırasındaki `clicks:deltas:proc:*` anahtarlarını da dahil edecek şekilde güncellendi.
- **SSRF, RBAC, Kota ve Link Durum Kuralları (`Backend/api`, `Backend/router`, `nginx`):**
  - IPv6 zone bypass'ı (`IsBlockedIP` & `IsInternalHost`), kısa/oktal/hex sayısal IPv4 biçimleri (`127.1`, `0x7f.0.0.1`) ve `PUBLIC_BASE_URL` üzerinde kendine yönlendirme (`IsSelfRedirect`) engellendi.
  - `AdminUpdateUser`, `AdminUpdateUserRole` ve `DeleteUser` içinde `GetRoles` hataları fail-closed hale getirildi; yetki denetimi paket muafiyeti kontrolünün önüne alındı; paket değişiminde eski/süresi dolmuş `plan_expires_at` temizlendi; `UserRepository.Delete` yalnızca `rows > 0` ise Redis/QR temizliği yapacak şekilde korundu.
  - Engelli/zararlı linklerin `isActive: true` olarak açılması engellendi ve `GetPlatformStats` aktif link sayacı `CountAdminLinks` ile hizalandı.
  - `CreateLink` kota denetimi PostgreSQL advisory lock ile atomikleştirildi; `Login` var olmayan e-postalarda sabit süreli Argon2id doğrulaması çalıştırır; `extractClientIP` yalnızca güvenilir proxy üzerinden gelen IP başlıklarını kabul eder.
  - `router` üzerinde `clicks` ve `max_clicks` kontrolü/artırımı atomik Lua betiği (`recordClickLua`) ve koşullu PostgreSQL artışı ile yarış durumuna kapatıldı; `GetLinkByAlias` `banned_until <= NOW()` kontrolünü SQL'e dahil etti; `WaitInfo`/`Continue` sonradan parolayla kilitlenen linkleri reddeder; `HEAD /{link}` tıklama saymaz; `oapi-codegen` parametre hataları JSON `ApiError` döner; Nginx'e güvenlik başlıkları ve servislere dinamik `Vary: Origin` CORS eklendi; OpenAPI şemaları güncellendi.
- **Frontend Tutarlılık ve Hata Düzeltmeleri (`Frontend/app`):**
  - `linkState.ts` ve `DecisionPanel.vue` içinde `blocked`/`malicious` önceliği router ile aynı sıraya alındı.
  - `useApi().blob` JSON hata yanıtlarını okuyarak `403 account_suspended` ve sunucu hata mesajlarını korur.
  - Oturum değişimi/kapanışı (`login`, `register`, `logout`, `deleteAccount`, `changePassword`, `endSuspendedSession`, 401 refresh düşüşü) için `mit_space_last` temizliği tamamlandı ve kullanılmayan `useAdmin` önbellekleri kaldırıldı.
  - `error.vue` SSR/istemci path uyuşmazlığı giderildi; `/w`, `/unlock`, `/error` rotalarında `encodeURIComponent` ve `?t=` aktarımları düzeltildi; `RegisterForm` e-posta kuralı `LoginForm` ile eşitlendi; ban süresi her açılışta anlık hesaplanır.

---

## 13. Hata Denetimi Düzeltmeleri, Özel Hata Sayfası ve Dışa Aktarma
- **Yeni:** Özel hata sayfası (`app/error.vue`, gerçek 404 durum koduyla). Linklerim ve link ziyaretleri için CSV / JSON dışa aktarma (ADR-015).
- **Güvenlik:**
  - Yöneticinin sahip ya da başka bir yöneticinin parolasını değiştirmesi engellendi.
  - İç ağ adreslerine kısa link verilmiyor ve router iç ağa istek atmıyor (SSRF).
  - Refresh token rotasyonu atomik hâle getirildi.
  - Nginx, sahtelenebilir ülke başlıklarını router'a iletmiyor.
- **Router:**
  - Bekleme bileti eklendi: bekleme süresi ve tıklama sınırı atlanamıyor, son geçerli ziyaretçi hedefe ulaşıyor.
  - Engelli linkler doğru ekranı gösteriyor.
  - Ölü hedef 30 saniye önbellekleniyor.
  - `GET /r/` hata ekranına gidiyor.
  - Sunucuya zaman aşımları ve düzgün kapanış eklendi.
- **Veri tutarlılığı:**
  - Tıklama farkları link kimliğiyle tutuluyor.
  - Güncellemede önbellek siliniyor; banlı kullanıcının linki yeniden açılmıyor.
  - Worker'da veri kaybı ve kapanış yarışı giderildi.
  - Süresi dolan askılar kendiliğinden kalkıyor.
  - Hedef değişince "güvenli" incelemesi sıfırlanıyor.
- **Arayüz:**
  - Ad değiştirmedeki "s" harfi hatası giderildi.
  - Askıya alınan oturum sebebiyle kapanıyor.
  - Çıkışta panel hafızası sıfırlanıyor.
  - Reklam anahtarı alt pakette kapatılabiliyor ve yalnızca bekleme sayfasında gösteriliyor.
  - Tıklama saatleri yerel saate çevriliyor.
  - Yönetim detayları hatayı gizlemiyor.
  - Üst menüden dönünce süzgeçler sıfırlanıyor.
- **Testler:** `Backend/api` ve `Backend/router` için birim testleri eklendi (`go test ./...`).

---

## 14. Kapsamlı Denetim Sonrası Hata ve Mantık Düzeltmeleri
- **Backend API (`Backend/api`, `openapi/api.yaml`):**
  - `AdminUpdateLink` içindeki `securityStatus`/`isActive` sıralama hatası giderildi; engelli linkin engeli kaldırılırken tek istekte yayına alınabilmesi ve engelli durumda `403 forbidden` dönmesi sağlandı.
  - Kendi parolasını değiştiren kullanıcılar için `ResetPasswordRequest.currentPassword` zorunlu hale getirildi (yanlış mevcut parolada `400 invalid_current_password`), hata ve başarı mesajları ayrıştırıldı.
  - `IsSelfRedirect` yol kanonikleştirmesi (`path.Clean`) ve `validateTargetURL` hatalı adres (`not-a-valid-url` → `ErrInvalidTargetURL`) ayrımı düzeltildi.
  - `csvSafe` formül koruması `\n` ve baştaki boşlukları kapsayacak şekilde güçlendirildi; refresh çerezi HTTPS'de `Secure` bayrağını dinamik alır.
  - Parola uzunluk denetimi UTF-8 rune sayısına çevrildi; `Register` ad doğrulaması, `UpdateLink`/`AdminUpdateLink` boş alan reddi, `waitSeconds` üst sınırı (`300`), `AdminListLinks` `status` enum denetimi, `ValidatePlanInput` fiyat üst sınırı ve JSON `ApiError` parametre hata yöneticisi eklendi.
  - `SetRole` işlem içine alındı, `UserRepository.Delete` hata kontrolü ve `CleanupExpiredAnonymousLinks` `reach:` temizliği düzeltildi; tıklama fark anahtarları `clicks:deltas:procs` kümesi üzerinden `O(1)` izlemeye alındı.
- **Backend Router & Altyapı (`Backend/router`, `openapi/router.yaml`, `docker-compose.yml`):**
  - `recordClickLua` `link:{alias}` yokken `max_clicks` atlamasını önlemek için PostgreSQL fallback'ine yönlendirildi; fallback hata kontrolü sıkılaştırıldı ve `pendingClicks` tam Redis `SCAN` yerine `SMEMBERS` ile `flushing` yarışı korumalı hale getirildi.
  - `checkTargetReachable` için `HEAD`/`GET` ayrı zaman aşımları, genişletilmiş `GET` (`Range: bytes=0-0`) fallback'i ve Cloudflare WAF-503 algılaması eklendi.
  - `POST /{link}/unlock` için Redis IP+alias hız sınırı (`429`), Argon2id eşzamanlılık semaforu (`8`), `64 KB` gövde sınırı ve `mit_unlock_ref` ile orijinal `Referer` koruması eklendi.
  - `/r/{alias}/` sonda eğik çizgi desteği, bilinmeyen alt yolların hata sayfasına yönlendirilmesi, `WaitInfo` boş `t` için `400` yanıtı ve geçici Redis hatalarının `ticket_invalid`'den ayrılması sağlandı; `postgres`/`redis` healthcheck'leri eklendi.
- **Frontend (`Frontend/app`):**
  - Yönetici karar paneli ve tablosunda engelli link anahtarı/menüsü backend kurallarıyla hizalandı; kullanıcı link ayarlarında engelli link ve paket dışı parola kontrolleri düzeltildi.
  - `useUnsavedGuard` oturum kapanış/yönlendirme çakışmaları ve `settings.vue` mevcut parola alanı eklendi.
  - Hata sayfasında çift `encodeURIComponent`, link detayda `Analiz` sekmesinin yeniden yüklenmesi, `Form.vue` `aliasState` sıfırlaması, `useApi.ts` `403 account_suspended` mesajı, SSR `auth.ts` yönlendirme koruması, `UiTape` sıfır/küçük kota gösterimi, admin kullanıcı detayı tarih/uyarı kontrolleri, anonim QR `@error` geri dönüşü ve mobil alt çubuk boşluğu düzeltildi.

## 15. Kullanıcının Paket Değiştirmesi ve Paket Düşürmede Askı
- **Backend:** `PUT /auth/user/plan` uç noktası, `00005_link_plan_grace.sql` migration'ı (`links.plan_grace_until`), `UpdateUserPlan` / `ApplyLinkPlanGrace` / `DeleteExpiredGraceLinks` sorguları, işlem içi `UserRepository.ChangePlan`, `AuthService.ChangeOwnPlan` ve worker'da `CleanupExpiredGraceLinks` eklendi. `LinkItem.planGraceUntil` alanı döner. `sqlc` ve `oapi-codegen` kodları yeniden üretildi.
- **Frontend:** Paketim sayfasına onay pencereli "Bu pakete geç" düğmeleri, `useAuth().changePlan` ve link durumunda "Askıda" (silinme tarihiyle) eklendi.

---

## 16. Yönetici Link İnceleme Sayfasına Grafiksel Analiz Eklenmesi
- **Frontend:** Yönetici link inceleme sayfasında ([admin/links/[id].vue](file:///e:/Projeler/makeitiny/Frontend/app/pages/dashboard/admin/links/%5Bid%5D.vue)) `Ziyaretler` sekmesindeki log tablosunun altına, yönetim panelinin mavi tema token'larıyla uyumlu grafiksel analiz bölümü (`LinksAnalysis` + `ChartBars`/`ChartShares`, en yeni 500 ziyaret üzerinden `"Tüm geçmiş"` kapsamıyla) eklendi.
- **Yenileme Senkronizasyonu:** [Visits.vue](file:///e:/Projeler/makeitiny/Frontend/app/components/links/Visits.vue) içindeki yenileme eylemi (`@refreshed`) ve sekme dönüşü, yönetici link özetini ve analiz grafiklerini birlikte günceller hale getirildi.

---

## 17. Yönetici Dışa Aktarma (CSV / JSON) ve Beyaz Makas Favicon
- **Backend (`openapi/api.yaml`, `Backend/api`):**
  - Filtreli tüm linkler (`GET /admin/links/export`), aramaya uyan tüm kullanıcılar (`GET /admin/users/export`) ve belirli bir kullanıcının tüm linkleri (`GET /admin/users/{id}/links/export`) için CSV (UTF-8 BOM + `csvSafe`) ve JSON dışa aktarma uç noktaları eklendi; `oapi-codegen` kodu ve handler rota/format testleri güncellendi.
- **Frontend (`Frontend/app`, `Frontend/public`):**
  - Yönetici linkler (`admin/links/index.vue`) ve kullanıcılar (`admin/users/index.vue`) sayfalarında arama kutusunun soluna indirme menüsü eklendi; kullanıcı detayı (`admin/users/[id].vue`) linkler bölümüne uyumlu `Dışa aktar` düğmesi konuldu.
  - Tarayıcı sekme ikonu için beyaz makas SVG (`Frontend/public/favicon.svg`) oluşturuldu ve `Frontend/nuxt.config.ts` `<head>` içine eklendi.

