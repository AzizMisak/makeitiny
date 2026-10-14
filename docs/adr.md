# Mimari Karar Kayıtları (Architecture Decision Records - ADR)

**Proje:** MakeItTiny  
**Tarih:** 2026-10-14  
**Durum:** Aktif / Yaşayan Doküman  

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
