# Mimari Karar Kayıtları (Architecture Decision Records - ADR)

**Proje:** MakeItTiny  
**Tarih:** 2026-10-14  
**Durum:** Aktif Doküman  

Bu doküman, MakeItTiny projesinin geliştirilme sürecinde alınan temel mimari kararları, bu kararların gerekçelerini ve sistem üzerindeki etkilerini kayıt altına almak amacıyla hazırlanmıştır.

---

## İçindekiler
1. [ADR-001: Monorepo Proje Yapısı](#adr-001-monorepo-proje-yapısı)
2. [ADR-002: Backend ve Router Servis Ayrımı](#adr-002-backend-ve-router-servis-ayrımı)
3. [ADR-003: Backend Dili Olarak Go ve Web Yaklaşımı](#adr-003-backend-dili-olarak-go-ve-web-yaklaşımı)
4. [ADR-004: Katmanlı Mimari (Layered Architecture)](#adr-004-katmanlı-mimari-layered-architecture)
5. [ADR-005: Spec-First Yaklaşımı ve OpenAPI (oapi-codegen)](#adr-005-spec-first-yaklaşımı-ve-openapi-oapi-codegen)
6. [ADR-006: Veri ve Önbellek Katmanı (PostgreSQL & Redis)](#adr-006-veri-ve-önbellek-katmanı-postgresql--redis)
7. [ADR-007: Frontend Teknolojisi (Nuxt & Tailwind CSS)](#adr-007-frontend-teknolojisi-nuxt--tailwind-css)
8. [ADR-008: Bağımsız Konteynerizasyon ve Swagger Dokümantasyonu](#adr-008-bağımsız-konteynerizasyon-ve-swagger-dokümantasyonu)
9. [ADR-009: Path Bazlı Nginx Yönlendirmesi ve Geliştirme/Prod Port Stratejisi](#adr-009-path-bazlı-nginx-yönlendirmesi-ve-geliştirmeprod-port-stratejisi)
10. [ADR-010: Kimlik Doğrulama, Şifreleme, Link Üretimi ve Redis Hash Mimarisi](#adr-010-kimlik-doğrulama-şifreleme-link-üretimi-ve-redis-hash-mimarisi)

---

### ADR-001: Monorepo Proje Yapısı

* **Durum:** Kabul Edildi
* **Bağlam:** Projede frontend, iki ayrı Go mikroservisi, OpenAPI şemaları ve Nginx reverse proxy yapılandırması bulunmaktadır.
* **Karar:** Tüm bileşenlerin tek bir repository (Monorepo) altında tutulmasına karar verildi.
* **Gerekçe:** Proje şu aşamada derli toplu ve kompakt bir yapıdadır. Birden fazla repository yönetimi, ayrı CI/CD süreçleri ve repolar arası bağımlılık takibi yerine tek repo üzerinden tüm projeyi görmek geliştirme hızını ve koordinasyonu artırır.

---

### ADR-002: Backend ve Router Servis Ayrımı

* **Durum:** Kabul Edildi
* **Bağlam:** URL kısaltma platformlarında yönetimsel işlemler (kullanıcı girişi, link oluşturma, analiz ekranları) ile bağlantı yönlendirme (`redirect`) trafiği çok farklı karakteristiklere sahiptir.
* **Karar:** Backend iki bağımsız Go servisine ayrıldı:
  1. **`api` Servisi:** Kullanıcı ile doğrudan etkileşime giren; kimlik doğrulama, link oluşturma, güncelleme ve veri okuma gibi iş kurallarını işleten yönetim servisi.
  2. **`router` Servisi:** Yalnızca kısaltılmış bağlantıları karşılayıp hedef URL'e en düşük gecikmeyle (`302 Found`) yönlendiren ve erişim kayıtlarını tutan yüksek performanslı servis.
* **Gerekçe:** Link yönlendirme trafiği ani tepe noktaları (spike) görebilir. Bu ayrım sayesinde yönlendirme altyapısı, yönetim API'sindeki olası yükten veya çökmelerden etkilenmez; bağımsız olarak ölçeklenebilir.

---

### ADR-003: Backend Dili Olarak Go ve Web Yaklaşımı

* **Durum:** Kabul Edildi
* **Bağlam:** Yüksek eşzamanlılık (concurrency), düşük bellek tüketimi ve hızlı derleme ihtiyacı.
* **Karar:** Backend servisleri için Go (Golang) seçildi.
* **Gerekçe:**
  * Go, geliştiricinin diğer projelerinden aşina olduğu, öğrenmesi ve bakımı kolay, derleme ve çalışma hızı yüksek bir dildir.
  * Web çatısı değerlendirmelerinde:
    * **Gin:** Öğrenmesi en kolay, topluluk desteği güçlü ve yüksek performanslı seçenek olarak değerlendirildi.
    * **Chi:** Görece daha fazla manuel yapılandırma gerektirdiği için elendi.
    * **Fiber:** Standart `net/http` ekosistemiyle tam uyumlu olmaması ve belirli senaryolarda kaynak optimizasyonu nedeniyle tercih edilmedi.
    * **Echo:** Ekosistemde Gin kadar yaygın kabul görmediği için elendi.
  * OpenAPI sözleşmesinden kod üreten `oapi-codegen` aracıyla Go standart kütüphanesinin (`net/http`) saf ve hafif gücünden doğrudan faydalanıldı.

---

### ADR-004: Katmanlı Mimari (Layered Architecture)

* **Durum:** Kabul Edildi
* **Bağlam:** Kod tabanının büyümesi durumunda spagetti kod oluşumunu önlemek ve sorumlulukları net şekilde ayırmak gerekmektedir.
* **Karar:** Servisler içinde `handler`, `service` ve `repository` katmanlarından oluşan katmanlı mimari benimsendi.
* **Gerekçe:** 
  * HTTP isteklerini karşılayan katman (`handler`), iş mantığını yürüten katman (`service`) ve veritabanı/cache işlemlerini yapan katman (`repository`) birbirinden ayrıldı.
  * Henüz tüm katman detayları kesinleşmemiş olsa da bu yapı, gelecekte yeni veritabanı sürücülerine geçmeyi, birim testleri yazmayı ve kodu genişletmeyi kolaylaştırır.

---

### ADR-005: Spec-First Yaklaşımı ve OpenAPI (oapi-codegen)

* **Durum:** Kabul Edildi
* **Bağlam:** Frontend ve Backend ekiplerinin/parçalarının ortak bir dilde buluşması ve uç nokta uyumsuzluklarının önlenmesi ihtiyacı.
* **Karar:** Kök dizindeki `openapi/` klasöründe `api.yaml` ve `router.yaml` şemaları oluşturuldu. Kodlar bu sözleşmeye göre üretildi (**Spec-First**).
* **Gerekçe:**
  * Önce sözleşme yazılarak hem ön yüz hem arka yüz aynı kurallara bağlandı.
  * Go tarafında `oapi-codegen` kullanılarak HTTP yönlendiricileri ve veri tipleri doğrudan YAML dosyasından üretildi; elle kod yazarken oluşabilecek insan hataları sıfırlandı.

---

### ADR-006: Veri ve Önbellek Katmanı (PostgreSQL & Redis)

* **Durum:** Kabul Edildi
* **Bağlam:** Kalıcı verilerin güvenliği ve yönlendirme hızının optimize edilmesi.
* **Karar:** İlişkisel veriler için PostgreSQL 16, önbellek ve anlık okuma için Redis 7 ayrı birer konteyner olarak yapılandırıldı.
* **Gerekçe:**
  * **PostgreSQL:** Kullanıcı hesapları, şifreler, bağlantı eşleşmeleri ve ilişkisel meta verilerin ACID garantisiyle saklanmasını sağlar.
  * **Redis:** Özellikle `router` servisinin link çözümlemelerini mikrosaniyeler mertebesinde yapabilmesi için sık kullanılan linkleri bellekte tutar; veritabanı üzerindeki okuma yükünü ortadan kaldırır.

---

### ADR-007: Frontend Teknolojisi (Nuxt & Tailwind CSS)

* **Durum:** Kabul Edildi
* **Bağlam:** Kullanıcı dostu, hızlı geliştirilebilen ve modern bir arayüz ihtiyacı.
* **Karar:** Frontend için **Nuxt** (Vue 3) ve stil kütüphanesi olarak **Tailwind CSS** seçildi.
* **Gerekçe:**
  * **Nuxt:** Geliştiricinin aşina olduğu, pratik, hazır routing ve state kabiliyetleri sunan güçlü bir framework'tür.
  * **Tailwind CSS:** Hızlı prototipleme, esnek utility sınıfları ve modern UI bileşenlerinin kolayca oluşturulabilmesi için tercih edildi.

---

### ADR-008: Bağımsız Konteynerizasyon ve Swagger Dokümantasyonu

* **Durum:** Kabul Edildi
* **Bağlam:** Tüm sistemin tek komutla çalıştırılabilmesi ve API uç noktalarının canlı olarak izlenebilmesi.
* **Karar:** Docker Compose ile modüler servis mimarisi kuruldu; bağımsız bir Swagger UI servisi eklendi.
* **Gerekçe:**
  * Her bir servis (`api`, `router`, `postgres`, `redis`, `nginx`, `swagger`) kendi izole ortamında çalışır.
  * Geliştirici veya kullanıcı `docker compose up -d` komutu ile tüm sistemi bağımlılıklarıyla birlikte ayağa kaldırabilir.
  * Swagger UI ayrı bir konteynerde `8082` portunda çalıştırılarak, Go servislerinin kodlarına dokunmadan `api` ve `router` sözleşmeleri tek bir açılır listeden (dropdown) interaktif olarak test edilebilir hale getirildi.

---

### ADR-009: Path Bazlı Nginx Yönlendirmesi ve Geliştirme/Prod Port Stratejisi

* **Durum:** Kabul Edildi (Güncelleme Notu)
* **Bağlam:** İlk tasarımda Nginx üzerinde alt alan adlarına (`api.domain.com`, `router.domain.com`) dayalı bir yönlendirme kurgulanmıştı. Ancak geliştirme aşamasında henüz aktif bir alan adının (domain) bulunmaması, alt alan adlarının kısaltılmış linkleri gereksiz uzatması, Nginx'te tanımsız `Host` başlıklarıyla gelen isteklerin ilk sıradaki `api` servisine düşmesi ve Nuxt/Swagger yapılandırmalarında port çakışmalarının yaşanması nedeniyle yönlendirme ve port stratejisinde revizyona gidildi.
* **Karar:**
  1. **Path (Yol) Bazlı Tekil Giriş Kapısı:** Subdomain ayrımı yerine tek bir sunucu bloğu (`localhost` / `127.0.0.1`) üzerinden path bazlı yönlendirmeye geçildi:
     * `/api/*` → Docker iç ağındaki `http://api:8080/` servisine yönlendirilir.
     * `/r/*` → Docker iç ağındaki `http://router:8081/` servisine yönlendirilir.
     * `/` (diğer tüm yollar) → Docker iç ağındaki `http://frontend:3000` servisine yönlendirilir.
  2. **Tanımsız Host ve Dış İsteklerin Engellenmesi:** Nginx'e `listen 80 default_server; server_name _;` bloğu eklenerek `localhost` ve `127.0.0.1` dışındaki tanımsız adreslerden gelen isteklerin `api` servisine düşmesi önlendi ve `444` koduyla bağlantıların doğrudan kapatılması sağlandı.
  3. **Geliştirme (Dev) ve Canlı (Prod) Port Politikası:**
     * Geliştirme sürecinde tüm sistem `docker compose up` ile çalıştırıldığından, hata ayıklama ve doğrudan erişim kolaylığı için tüm servislerin portları (`80`, `3000`, `8080`, `8081`, `8082`, `5432`, `6379`) host makineye açık bırakıldı.
     * Buna karşın Nginx, servislerle her zaman Docker iç ağı üzerinden (`api:8080`, `router:8081`, `frontend:3000`) haberleşecek şekilde tutuldu. Prod ortamına çıkılırken diğer servislerin dış portları kapatılıp yalnızca Nginx dış dünyaya açık bırakılacaktır.
  4. **Frontend ve Swagger Uyumlaması:**
     * Nuxt `runtimeConfig.public` altındaki `apiBase` (`http://localhost:3000` çakışması giderilerek `http://localhost/api`) ve `routerBase` (`http://localhost/r`) adresleri yeni Nginx path yapısına göre güncellendi.
     * OpenAPI şemalarına (`openapi/api.yaml` ve `openapi/router.yaml`) hem Nginx proxy adresleri (`http://localhost/api`, `http://localhost/r`) hem de doğrudan geliştirme portları (`http://localhost:8080`, `http://localhost:8081`) eklendi; `router` servisinin CORS izin listesi yerel tarayıcı ve frontend origin'lerini kapsayacak şekilde genişletildi.
* **Gerekçe:**
  * `Host` başlığı karmaşası ve yerel geliştirme ortamında `/etc/hosts` düzenleme zorunluluğu tamamen ortadan kaldırıldı.
  * Kısaltılmış linkler için `/r/` öneki kullanılarak Frontend sayfa rotaları ile kısa linklerin çakışma ihtimali sıfırlandı.
  * Geliştirme ortamında esneklik korunurken, canlı ortama geçişte yalnızca portların kapatılmasıyla güvenli çalışacak mimari garanti altına alındı.

---

### ADR-010: Kimlik Doğrulama, Şifreleme, Link Üretimi ve Redis Hash Mimarisi

* **Durum:** Kabul Edildi
* **Bağlam:** Altyapı ve veritabanı şemalarının tamamlanmasının ardından; kimlik doğrulama yöntemi, parola saklama güvenliği, kısa link (`alias`) üretim algoritması, Redis önbellekleme veri modeli ve şifreli link yönlendirme akışının netleştirilmesi gerekti.
* **Karar:**
  1. **Kimlik Doğrulama (JWT):** Oturum yönetimi için durumsuz (stateless) **JWT (`HS256` - Bearer Token)** kullanılmasına karar verildi. Kayıt sırasında `name` alanı gönderilmezse varsayılan kullanıcı adı olarak e-posta adresinin `@` işaretinden önceki kısmı atanacaktır.
  2. **Parola Güvenliği (`Argon2id`):** Hem kullanıcı hesap parolaları (`users.password_hash`) hem de şifreli kısa link parolaları (`links.password_hash`) Go'nun `golang.org/x/crypto/argon2` paketiyle **Argon2id** algoritması kullanılarak hash'lenecektir.
  3. **Kısa Link Üretim Algoritması (Rastgele Base62):** Kullanıcı özel bir takma ad (`alias`) belirtmediğinde, kriptografik olarak rastgele (`crypto/rand`) **Base62 (`[0-9a-zA-Z]`)** karakter kümesinden kısa kod üretilecektir.
  4. **Veritabanı (`goose` + `sqlc`) ve Redis (`go-redis` Hash) Veri Modeli:**
     * Veritabanı şema göçleri ortak `Backend/db/migrations` dizininde **Goose** ile, tip güvenli sorgu üretimi ise `api` ve `router` servislerinde birbirinden bağımsız **sqlc** yapılandırmalarıyla yönetilecektir.
     * Redis üzerinde veriler bir NoSQL tablosu/dokümanı gibi **Redis Hash (`HSET` / `HGETALL` / `HINCRBY`)** yapısında (`link:{alias}` anahtarı altında `id`, `user_id`, `target_url`, `password_hash`, `clicks` alanlarıyla) tutulacaktır.
     * Yönlendirme ve veri değişikliklerinde önce PostgreSQL işlemi gerçekleştirilecek, hemen ardından Redis Hash kaydı güncellenecektir (`HINCRBY` / `HSET` / `DEL`).
  5. **Şifreli Link Yönlendirme Akışı:** Bir kullanıcı tarayıcıdan şifreli bir kısa linke (`GET /r/{link}`) eriştiğinde, `router` servisi kullanıcıyı Frontend üzerindeki şifre çözme sayfasına (`/unlock/{link}`) yönlendirecek; parola doğrulaması `POST /r/{link}/unlock` üzerinden tamamlanacaktır.
* **Gerekçe:**
  * Redis Hash kullanımı sayesinde tüm nesneyi serialize/deserialize etmeye gerek kalmadan tekil alanlar (`clicks`, `target_url`) atomik olarak güncellenebilir.
  * Önce PostgreSQL ardından Redis güncellemesi yapılarak kalıcı veri tutarlılığı (consistency) garanti altına alınır.
  * Argon2id ve kriptografik Base62 ile hem parola güvenliği hem de kısa linklerin tahmin edilemezliği sağlanır.
