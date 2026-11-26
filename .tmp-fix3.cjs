const fs = require('fs')
const p = 'Backend/api/internal/worker/worker.go'
let s = fs.readFileSync(p, 'utf8')
const a = `	procKey := fmt.Sprintf("%s%d", ClickProcPrefix, time.Now().UnixNano())
	if err := w.rdb.Rename(ctx, ClickDeltasKey, procKey).Err(); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "no such key") {
			log.Printf("worker click deltas RENAME uyarısı: %v", err)
		}
		return
	}
	_ = w.rdb.SAdd(ctx, clickDeltasProcsKey, procKey).Err()`
const b = `	procKey := fmt.Sprintf("%s%d", ClickProcPrefix, time.Now().UnixNano())
	// Kopya, adı değişmeden önce izleme kümesine yazılır. Böylece router, RENAME ile SADD arasındaki
	// anda bekleyen tıklamaları gözden kaçırıp tıklama sınırını aşacak eksik bir sayı önbelleklemez.
	_ = w.rdb.SAdd(ctx, clickDeltasProcsKey, procKey).Err()
	if err := w.rdb.Rename(ctx, ClickDeltasKey, procKey).Err(); err != nil {
		_ = w.rdb.SRem(ctx, clickDeltasProcsKey, procKey).Err()
		if !strings.Contains(strings.ToLower(err.Error()), "no such key") {
			log.Printf("worker click deltas RENAME uyarısı: %v", err)
		}
		return
	}`
if (!s.includes(a)) throw new Error('MISS worker')
s = s.replace(a, b)
fs.writeFileSync(p, s)
console.log('fix3 ok')
