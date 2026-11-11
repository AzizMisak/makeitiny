# MakeItTiny

**MakeItTiny**, yüksek performanslı ve mikroservis mimarisine sahip bir URL kısaltma ve link yönetim platformudur.

Proje; yönetim işlemleri (kimlik doğrulama, link oluşturma ve analitik) ile son kullanıcı yönlendirme (redirect) trafiğini birbirinden izole eden iki ayrı Go servisi üzerine kurulmuştur. API sözleşmeleri **Spec-First** (OpenAPI 3.0) yaklaşımıyla tasarlanmış olup tüm altyapı Docker Compose ile orkestre edilmektedir. Detaylı teknik kararlar için [Mimari Karar Kayıtları (ADR)](docs/adr.md) dokümanını inceleyebilirsiniz.

---

## Konteyner Yapısı ve Servisler

Proje bünyesinde yer alan servislerin port ve erişim özeti (Geliştirme sürecinde doğrudan test ve hata ayıklama kolaylığı için tüm portlar dışarıya açıktır; canlı/prod ortamına geçilirken yalnızca `nginx` portu dışarıya açık bırakılacaktır):

| Konteyner Adı | Kaynak / İmaj | Host Portu (Dev) | Dahili Port | Görevi |
| :--- | :--- | :--- | :--- | :--- |
| **nginx** | `nginx:1.27-alpine` | `80` | `80` | Reverse proxy ve path (yol) bazlı trafik yönlendirme |
| **frontend** | `./Frontend` (Nuxt 3) | `3000` | `3000` | Web kullanıcı arayüzü (Nuxt 3 & Tailwind CSS) |
| **swagger** | `swaggerapi/swagger-ui` | `8082` | `8080` | OpenAPI şemalarını gösteren dokümantasyon arayüzü |
| **api** | `./Backend/api` (Go) | `8080` | `8080` | Kimlik doğrulama ve link yönetim API servisi |
| **router** | `./Backend/router` (Go) | `8081` | `8081` | Kısaltılmış linklerin yüksek hızlı yönlendirme servisi |
| **postgres** | `postgres:16-alpine` | `5432` | `5432` | İlişkisel veritabanı (kullanıcı, link ve meta veriler) |
| **redis** | `redis:7-alpine` | `6379` | `6379` | Link çözümleme ve önbellekleme (cache) katmanı |

---

### 1. Nginx (`nginx`)
- **Port:** `80:80` (Dışarıya açık)
- **Erişim:** Yalnızca `http://localhost` ve `http://127.0.0.1` (Tanımsız `Host` başlığı veya dış adreslerden gelen istekler `default_server` bloğu ile `444` döndürülerek engellenir).
- **Sorumluluk:**
  Sistemin ön yüzündeki ters vekil (reverse proxy) sunucusudur. Tüm servislerle Docker iç ağı üzerinden haberleşir ve gelen HTTP isteklerini **path (yol)** bazlı olarak ilgili iç servise iletir:
  - `http://localhost/api/*` üzerinden gelen istekleri `http://api:8080/*` yönetim servisine yönlendirir.
  - `http://localhost/r/*` üzerinden gelen kısaltılmış link isteklerini `http://router:8081/*` yönlendirme servisine iletir.
  - `http://localhost/` (geriye kalan tüm yollar) üzerinden gelen istekleri `http://frontend:3000` web arayüzüne yönlendirir.

---

### 2. Swagger UI (`swagger`)
- **Port:** `8082:8080` (Dışarıya açık: `http://localhost:8082`)
- **Sorumluluk:**
  Projenin API dokümantasyon merkezidir. Doğrudan `8082` portundan hizmet verir.
  - `./openapi` dizinindeki sözleşmeleri (`api.yaml` ve `router.yaml`) okur.
  - Sağ üst köşesindeki açılır menü (dropdown) üzerinden **API Service** ve **Router Service** dokümantasyonları arasında geçiş yapma, sunucu listesinden (Servers) ister Nginx Reverse Proxy (`http://localhost/api`, `http://localhost/r`) ister doğrudan servis portları (`http://localhost:8080`, `http://localhost:8081`) üzerinden uç noktaları interaktif olarak test etme imkânı sunar.

---

### 3. API Servisi (`api`)
- **Port:** `8080:8080` (Nginx üzerinden `http://localhost/api` ile veya geliştirme ortamında doğrudan `http://localhost:8080` ile erişilir; prod ortamında dış dünyaya kapatılır)
- **Sorumluluk:**
  Yönetimsel ve iş mantığı süreçlerini yürüten ana RESTful backend servisidir.
  - Kullanıcı kayıt (`/auth/register`), giriş (`/auth/login`) ve şifre sıfırlama işlemleri.
  - Kısaltılmış link oluşturma, güncelleme, silme ve kullanıcıya ait linklerin listelenmesi.
  - Veri okuma ve yazma işlemlerinde PostgreSQL ve Redis ile haberleşir.

---

### 4. Router Servisi (`router`)
- **Port:** `8081:8081` (Nginx üzerinden `http://localhost/r/{link}` ile veya geliştirme ortamında doğrudan `http://localhost:8081/{link}` ile erişilir; prod ortamında dış dünyaya kapatılır)
- **Sorumluluk:**
  Kısaltılmış bağlantıların çözümlenmesi ve son kullanıcıya yönlendirilmesinden sorumlu, hafif ve yüksek performanslı yönlendirme servisidir.
  - `/{link}` (Nginx üzerinden `/r/{link}`): İlgili linki Redis/PostgreSQL üzerinden bularak `302 Found` durum koduyla hedef adrese yönlendirir (analitik kaybını önlemek için tarayıcı önbelleklemesi devre dışı bırakılır).
  - `/{link}/unlock`: Şifreli linklerin kilidini açar ve yönlendirmeyi tamamlar.
  - `/{link}/preview`: Linkin hedef adresini önizleme amacıyla görüntüler.
  - `/health`: Servis sağlık kontrolünü sağlar.

---

### 5. PostgreSQL (`postgres`)
- **Port:** `5432:5432` (Geliştirme ortamında dışarıya açık: `localhost:5432`)
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
- **Port:** `6379:6379` (Geliştirme ortamında dışarıya açık: `localhost:6379`, prod ortamında yalnızca Docker iç ağı)
- **Sorumluluk:**
  Özellikle yönlendirme (`router`) servisinin link hedeflerine mikrosaniyeler mertebesinde erişebilmesi için kullanılan bellek içi (in-memory) önbellekleme katmanıdır. Sık erişilen link verilerini hafızada tutarak veritabanı yükünü minimize eder.

---

### 7. Frontend (`frontend`)
- **Port:** `3000:3000` (Nginx üzerinden: `http://localhost` veya geliştirme ortamında doğrudan: `http://localhost:3000`)
- **Sorumluluk:**
  Kullanıcıların link oluşturabileceği, kısaltılmış bağlantıları yönetebileceği ve erişim istatistiklerini takip edebileceği Nuxt 3 ve Tailwind CSS ile geliştirilmiş modern web kullanıcı arayüzüdür. API isteklerini `http://localhost/api`, yönlendirme bağlantılarını ise `http://localhost/r` taban adresleri üzerinden gerçekleştirir.

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
