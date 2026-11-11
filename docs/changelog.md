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
