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
