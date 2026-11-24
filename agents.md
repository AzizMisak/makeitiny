# AGENTS.md — MakeItTiny Yapay Zeka Asistanı Rehberi ve Kuralları

Bu doküman, **MakeItTiny** projesi üzerinde çalışan yapay zeka asistanlarının (AI Agents) uyması gereken kesin çalışma kurallarını, geliştirme kültürünü, proje mimarisini ve işleyiş yolunu tanımlar.

---

## 1. Kesin Çalışma ve İletişim Kuralları

1. **İstenmeyen Hiçbir Şey Yapılmaz:**
   - Bu proje bir link kısaltma web projesi ödevidir.
   - Kullanıcının açıkça **"yap"** demediği hiçbir özelliği ekleme, hiçbir dosyayı değiştirme veya silme. Kendi inisiyatifinle kapsam dışı kod yazma.
2. **Belirsizliklerde Önce Tartışma:**
   - Eğer kullanıcının isteğinde belirsiz bir nokta veya birden fazla çözüm yolu varsa, doğrudan kod yazmaya başlama.
   - Önce durumu analiz et, olası seçenekleri artıları ve eksileriyle birlikte açıkla ve kullanıcının kararını bekle.
3. **Commit Yasağı:**
   - **Asla `git commit` veya `git push` çalıştırma.** Versiyon kontrolü ve commit işlemleri yalnızca kullanıcı tarafından yapılır.
4. **Tur Sonu Özeti Zorunluluğu:**
   - Her turun sonunda o turda neler yaptığını, hangi dosyaları okuduğunu veya değiştirdiğini maddeler halinde, net ve anlaşılır biçimde kullanıcıya anlat.
5. **Dokümantasyon Güncelliği:**
   - Alınan yeni mimari kararları, önceki kayıtları silmeden `docs/adr.md` dosyasına yeni bir ADR maddesi olarak ekle.
   - Tamamlanan işleri çok uzatmadan özlü şekilde `docs/changelog.md` dosyasına kaydet.
6. **Proje Dizini Dışına Erişim ve Dosya Yazma Yasağı:**
   - Proje kök dizini dışındaki hiçbir klasöre veya dosyaya erişme, değişiklik yapma veya yeni dosya/klasör oluşturma.
   - Windows ortamında `/tmp/...` gibi Unix tarzı yollar sürücü kökünde (ör. `E:\tmp`) klasör oluşturduğu için derleme veya test doğrulamalarında asla `/tmp/...` gibi yollar kullanma; dosya çıktısı üretmeyen komutları (`go vet ./...`, `go build -o NUL .` vb.) tercih et.

---

## 2. Projenin Amacı ve Vizyonu

**MakeItTiny**, uzun URL'leri kısa ve paylaşılabilir bağlantılara (`localhost/r/{alias}`) dönüştüren, tıklama sayılarını takip eden ve isteğe bağlı parola koruması sunan, mikroservis mimarisine sahip bir URL kısaltma ve link yönetim platformudur.

- **Tasarım Felsefesi:** Gereksiz karmaşadan (over-engineering) kaçınılır. Sistem; bir kullanıcının rahatça **kayıt olup giriş yapabileceği, şifresini değiştirebileceği, hesabını silebileceği, kısa link oluşturup görüntüleyebileceği, link detaylarını düzenleyebileceği (hedef adres, takma ad, şifre ekleme/kaldırma) ve silebileceği** yalın ve sağlam bir akış üzerine kuruludur.
- **Gelecek Yol Haritası (`notes.txt`):** Temel sistemin üzerine ilerleyen aşamalarda detaylı erişim analitiği (`/analysis`), QR kod üretimi, güvenli link testi (`/securetest`) ve reklam/ödeme entegrasyonları gibi modüller yalnızca kullanıcı talep ettiğinde eklenecektir.

---

## 3. Mimari Yapı ve Teknoloji Yığını (Monorepo)

Tüm sistem kök dizindeki `docker-compose.yml` ile orkestre edilir ve geliştirme sürecinde `docker compose up` komutuyla çalıştırılır:

| Servis | Dizin / Kaynak | Port (Dev) | Sorumluluk ve Teknolojiler |
| :--- | :--- | :--- | :--- |
| **nginx** | `./nginx/nginx.conf` | `80:80` | Ters vekil (Reverse Proxy). Tüm servislerle Docker iç ağı üzerinden konuşur. Path bazlı yönlendirme yapar. |
| **frontend** | `./Frontend` | `3000:3000` | Nuxt 3 (Vue 3) & Tailwind CSS web arayüzü. |
| **api** | `./Backend/api` | `8080:8080` | Go 1.26 (`net/http`). Kullanıcı kimlik doğrulama (`Auth`) ve kısa link yönetim (`Links`) REST API servisi. |
| **router** | `./Backend/router` | `8081:8081` | Go 1.26 (`net/http`). Yüksek hızlı kısa link yönlendirme (`302 Found`), şifre çözme ve önizleme servisi. |
| **postgres** | `postgres:16-alpine` | `5432:5432` | Kalıcı ilişkisel veritabanı (`users` ve `links` tabloları). Veriler `./Backend/db/data` altında tutulur. |
| **redis** | `redis:7-alpine` | `6379:6379` | Bellek içi önbellek katmanı. Link verileri `link:{alias}` anahtarıyla **Redis Hash** yapısında tutulur. |
| **swagger** | `swaggerapi/swagger-ui` | `8082:8080` | `./openapi/api.yaml` ve `./openapi/router.yaml` şemalarını sunan interaktif API test arayüzü. |

### Yönlendirme (Routing) ve Port Kuralları
- **Path Bazlı Tek Giriş Kapısı (`nginx`):**
  - `http://localhost/api/*` → `http://api:8080/*`
  - `http://localhost/r/*` → `http://router:8081/*`
  - `http://localhost/*` → `http://frontend:3000`
- **Host Güvenliği:** Geliştirme ortamında Nginx yalnızca `localhost` ve `127.0.0.1` üzerinden gelen isteklere izin verir; tanımsız `Host` başlıklarını `default_server` bloğu (`return 444`) ile reddeder.
- **Dev vs. Prod Port Politikası:** Geliştirme aşamasında hata ayıklama kolaylığı için tüm portlar dışarıya açıktır. Prod ortamına geçilirken yalnızca `nginx` portu dış dünyaya açık bırakılır.
- **Nginx Yeniden Başlatma Notu:** `api`, `router` veya `frontend` konteynerleri yeniden derlenip oluşturulduğunda (`Recreate`), Nginx'in yeni konteyner IP'lerini alması için `docker compose restart nginx` çalıştırılmalıdır.

---

## 4. Teknik Standartlar ve İş Kuralları

1. **Spec-First & OpenAPI (`oapi-codegen`):**
   - API sözleşmeleri `./openapi/api.yaml` ve `./openapi/router.yaml` dosyalarında tanımlıdır.
   - Uç nokta veya şema değişikliklerinde önce YAML dosyaları güncellenir, ardından `Backend/api` ve `Backend/router` dizinlerinde `go generate ./...` çalıştırılarak `internal/gen/server.gen.go` dosyaları üretilir.
2. **Katmanlı Mimari (Layered Architecture):**
   - Go servisleri `internal/handler` (HTTP katmanı), `internal/service` (iş kuralları), `internal/repository` (veri erişim ve Redis senkronizasyonu) ve `internal/db` (`sqlc` tarafından üretilen tip güvenli sorgular) katmanlarından oluşur.
3. **Veritabanı (`goose` + `sqlc` + `pgx/v5`):**
   - Migration dosyaları ortak `./Backend/db/migrations` dizininde tutulur (`api` servisi başlangıçta otomatik `goose.Up` çalıştırır).
   - `sqlc` yapılandırmaları ise her iki servisin yalnızca kendi sorgularını barındırması için `Backend/api/sqlc.yaml` ve `Backend/router/sqlc.yaml` olarak ayrıdır.
4. **Redis Hash (`go-redis/v9`) ve Senkronizasyon Sırası:**
   - Kısa linkler Redis'te bir NoSQL tablosu/dokümanı gibi `link:{alias}` anahtarı altında **Redis Hash (`HSET`, `HGETALL`, `HINCRBY`)** olarak saklanır (`id`, `user_id`, `alias`, `target_url`, `password_hash`, `clicks` alanlarıyla).
   - **Kritik Kural:** Tüm veri oluşturma, güncelleme, silme ve tıklama artırma işlemlerinde **önce PostgreSQL işlemi** yapılır, **hemen ardından Redis Hash** güncellenir veya silinir.
5. **Kimlik Doğrulama ve Güvenlik (`JWT` & `Argon2id`):**
   - Oturum yönetimi stateless **JWT (`HS256` - Bearer Token)** ile yapılır.
   - Hem kullanıcı parolaları hem de şifreli link parolaları **`Argon2id`** (`golang.org/x/crypto/argon2`) algoritmasıyla PHC formatında hash'lenir.
   - Kayıt sırasında kullanıcı adı (`name`) girilmezse e-posta adresinin `@` işaretinden önceki kısmı varsayılan kullanıcı adı olarak atanır. Kullanıcı parolası en az 4 karakter olmalıdır.
6. **Kısa Link (`alias`) ve URL Kuralları:**
   - Kullanıcı özel bir `alias` belirtmezse kriptografik rastgele (`crypto/rand`) 6 karakterli **Base62 (`[0-9a-zA-Z]`)** kod üretilir.
   - Hedef URL başında `http://` veya `https://` belirtilmemişse hem Frontend hem Backend tarafında varsayılan olarak güvenli **`https://`** eklenir.
   - Şifreli bir kısa linke (`GET /r/{link}`) tarayıcıdan girildiğinde `router` servisi kullanıcıyı Frontend'deki `/unlock/{link}` sayfasına yönlendirir.
