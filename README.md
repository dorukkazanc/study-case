# Notification Service

SMS, Email ve Push kanalları üzerinden bildirim gönderen asenkron bir servis.

## Kurulum

```bash
docker compose up --build
```

## API

| Method | Endpoint | Açıklama |
|--------|----------|----------|
| POST | `/api/v1/notifications` | Bildirim oluştur |
| POST | `/api/v1/notifications/batch` | Toplu bildirim oluştur (max 1000) |
| GET | `/api/v1/notifications` | Listele |
| GET | `/api/v1/notifications/:id` | ID ile getir |
| DELETE | `/api/v1/notifications/:id` | İptal et |
| GET | `/api/v1/batches/:batchID` | Batch durumu |
| GET | `/health` | Health check |
| GET | `/metrics` | Prometheus metrikleri |
| GET | `/swagger/index.html` | API dokümantasyonu |

### Örnek İstek

```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": "+905551234567",
    "channel": "sms",
    "content": "Doğrulama kodunuz: 123456",
    "priority": "high"
  }'
```

## Testler

```bash
go test ./...
```

Mock tabanlı unit testler mevcut (`service`, `worker`). Repo, Redis ve goroutine/concurrency testleri integration gerektirdiği için eklemedim.

## Nasıl Çalışır

1. İstek gelir, notification DB'ye kaydedilir ve Redis priority queue'ya eklenir
2. Worker queue'dan çeker, provider'a (webhook) gönderir
3. Başarısızsa exponential backoff ile yeniden dener (5s → 10s → 20s...)
4. Max retry'a ulaşırsa dead letter queue'ya taşır

## Teknik Detaylar

- **Priority queue:** Redis Sorted Set — high=1, normal=2, low=3
- **Rate limiting:** Kanal başına 100 msg/sec
- **Max retry:** 3 (config ile değiştirilebilir)
- **Correlation ID:** Her istek `X-Request-Id` header'ı ile takip edilir

## Ortam Değişkenleri

| Değişken | Açıklama |
|----------|----------|
| `DATABASE_DSN` | PostgreSQL bağlantı string'i |
| `REDIS_ADDR` | Redis adresi |
| `PROVIDER_WEBHOOK_URL` | Bildirim gönderilecek webhook URL |
| `PROVIDER_RATE_LIMIT_RPS` | Saniyedeki max istek sayısı |
| `LOG_LEVEL` | Log seviyesi (debug/info/warn/error) |
