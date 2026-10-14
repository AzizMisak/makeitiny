# MakeItTiny

**MakeItTiny**, yüksek performanslı ve mikroservis mimarisine sahip bir URL kısaltma ve link yönetim platformudur.

Proje; yönetim işlemleri (kimlik doğrulama, link oluşturma ve analitik) ile son kullanıcı yönlendirme (redirect) trafiğini birbirinden izole eden iki ayrı Go servisi üzerine kurulmuştur. API sözleşmeleri **Spec-First** (OpenAPI 3.0) yaklaşımıyla tasarlanmış olup tüm altyapı Docker Compose ile orkestre edilmektedir. Detaylı teknik kararlar için [Mimari Karar Kayıtları (ADR)](docs/adr.md) dokümanını inceleyebilirsiniz.

---

## Konteyner Yapısı ve Servisler

Proje bünyesinde yer alan servislerin port ve erişim özeti:

| Konteyner Adı | Kaynak / İmaj | Host Portu | Dahili Port | Görevi |
| :--- | :--- | :--- | :--- | :--- |
| **nginx** | `nginx:1.27-alpine` | `80` | `80` | Reverse proxy ve domain bazlı trafik yönlendirme |
| **frontend** | `./Frontend` (Nuxt 3) | `3000` | `3000` | Web kullanıcı arayüzü (Nuxt 3 & Tailwind CSS) |
| **swagger** | `swaggerapi/swagger-ui` | `8082` | `8080` | OpenAPI şemalarını gösteren dokümantasyon arayüzü |
| **api** | `./Backend/api` (Go) | - | `8080` | Kimlik doğrulama ve link yönetim API servisi |
| **router** | `./Backend/router` (Go) | - | `8081` | Kısaltılmış linklerin yüksek hızlı yönlendirme servisi |
| **postgres** | `postgres:16-alpine` | `5432` | `5432` | İlişkisel veritabanı (kullanıcı, link ve meta veriler) |
| **redis** | `redis:7-alpine` | - | `6379` | Link çözümleme ve önbellekleme (cache) katmanı |

---

### 1. Nginx (`nginx`)
- **Port:** `80:80` (Dışarıya açık)
- **Erişim:** `http://localhost` (Frontend), `http://api.domain.com` ve `http://router.domain.com`
- **Sorumluluk:**
  Sistemin ön yüzündeki ters vekil (reverse proxy) sunucusudur. Gelen HTTP isteklerinin `Host` başlığını inceleyerek trafiği ilgili iç servise iletir:
  - `localhost` üzerinden gelen istekleri doğrudan `frontend:3000` web arayüzüne yönlendirir.
  - `api.domain.com` isteklerini doğrudan `api:8080` servisine yönlendirir.
  - `router.domain.com` isteklerini doğrudan `router:8081` servisine yönlendirir.

---

### 2. Swagger UI (`swagger`)
- **Port:** `8082:8080` (Dışarıya açık: `http://localhost:8082`)
- **Sorumluluk:**
  Projenin API dokümantasyon merkezidir. Domain veya proxy ayarlarından bağımsız olarak doğrudan `8082` portundan hizmet verir.
  - `./openapi` dizinindeki sözleşmeleri (`api.yaml` ve `router.yaml`) okur.
  - Sağ üst köşesindeki açılır menü (dropdown) üzerinden **API Service** ve **Router Service** dokümantasyonları arasında geçiş yapma ve uç noktaları doğrudan test etme imkânı sunar.

---

### 3. API Servisi (`api`)
- **Port:** `8080` (Yalnızca Docker iç ağı; Nginx üzerinden `api.domain.com` ile erişilir)
- **Sorumluluk:**
  Yönetimsel ve iş mantığı süreçlerini yürüten ana RESTful backend servisidir.
  - Kullanıcı kayıt (`/auth/register`), giriş (`/auth/login`) ve şifre sıfırlama işlemleri.
  - Kısaltılmış link oluşturma, güncelleme, silme ve kullanıcıya ait linklerin listelenmesi.
  - Veri okuma ve yazma işlemlerinde PostgreSQL ve Redis ile haberleşir.

---

### 4. Router Servisi (`router`)
- **Port:** `8081` (Yalnızca Docker iç ağı; Nginx üzerinden `router.domain.com` ile erişilir)
- **Sorumluluk:**
  Kısaltılmış bağlantıların çözümlenmesi ve son kullanıcıya yönlendirilmesinden sorumlu, hafif ve yüksek performanslı yönlendirme servisidir.
  - `/{link}`: İlgili linki Redis/PostgreSQL üzerinden bularak `302 Found` durum koduyla hedef adrese yönlendirir (analitik kaybını önlemek için tarayıcı önbelleklemesi devre dışı bırakılır).
  - `/{link}/unlock`: Şifreli linklerin kilidini açar ve yönlendirmeyi tamamlar.
  - `/{link}/preview`: Linkin hedef adresini önizleme amacıyla görüntüler.
  - `/health`: Servis sağlık kontrolünü sağlar.

---

### 5. PostgreSQL (`postgres`)
- **Port:** `5432:5432` (Dışarıya açık: `localhost:5432`)
- **Sorumluluk:**
  Kullanıcı hesapları, bağlantı bilgileri, şifreli link anahtarları ve kalıcı kayıtların saklandığı ilişkisel veritabanıdır. Veriler `./Backend/db/data` dizininde kalıcı (persistent volume) olarak tutulur.
- **DSN (Bağlantı Adresi) Kuralı:**
  Veritabanına erişim için genel DSN standardı:
  ```text
  postgres://<POSTGRES_USER>:<POSTGRES_PASSWORD>@<HOST>:<PORT>/<POSTGRES_DB>?sslmode=disable
  ```
  - **Host / Yerel Makineden Erişim (DBeaver, DataGrip, psql vb.):**
    ```text
    postgres://makeitiny:makeitiny@localhost:5432/makeitiny?sslmode=disable
    ```
  - **Docker Ağı İçinden Erişim (Konteynerler arası `DATABASE_URL`):**
    ```text
    postgres://makeitiny:makeitiny@postgres:5432/makeitiny?sslmode=disable
    ```

---

### 6. Redis (`redis`)
- **Port:** `6379` (Yalnızca Docker iç ağı)
- **Sorumluluk:**
  Özellikle yönlendirme (`router`) servisinin link hedeflerine mikrosaniyeler mertebesinde erişebilmesi için kullanılan bellek içi (in-memory) önbellekleme katmanıdır. Sık erişilen link verilerini hafızada tutarak veritabanı yükünü minimize eder.

---

### 7. Frontend (`frontend`)
- **Port:** `3000:3000` (Doğrudan: `http://localhost:3000` veya Nginx üzerinden: `http://localhost`)
- **Sorumluluk:**
  Kullanıcıların link oluşturabileceği, kısaltılmış bağlantıları yönetebileceği ve erişim istatistiklerini takip edebileceği Nuxt 3 ve Tailwind CSS ile geliştirilmiş modern web kullanıcı arayüzüdür.

---

## Ortam Değişkenleri (.env)

Veritabanı kullanıcı adı, şifresi ve veritabanı adı kök dizindeki `.env` dosyasından okunur. Yeni bir ortam kurulurken `.env.example` şablonundan faydalanabilirsiniz:

```bash
cp .env.example .env
```

---

## Çalıştırma

Tüm altyapıyı arka planda başlatmak için proje kök dizininde şu komutu çalıştırmanız yeterlidir:

```bash
docker compose up -d
```
