# Golang Module Jaeger

Modul tracing Jaeger untuk Go dengan **clean architecture**, mudah dibaca dan dipakai di berbagai layanan.

- **Process/span**: `server_ip` dan `exectime` (durasi span) ada di process/span Jaeger; `trace_id` dan `transaction_id` otomatis di tag.
- **API**: `GetTracing().Operation(ctx, ...)` atau `GetTracingForService("service-name").Operation(ctx, ...)` → `span.Tag()`, `span.Info/Error/Warning/Debug`, `defer span.Finish()`.
- **Multi-service (satu process)**: `ServiceNames` di config + `GetTracingForService(serviceName)` agar satu aplikasi muncul sebagai beberapa service di Jaeger (cocok untuk config dari DB/env).

## Field di Jaeger

| Lokasi | Field | Keterangan |
|--------|--------|-------------|
| **Process** | `server.ip` | IP server (dari config atau auto-detect) |
| **Span (tag)** | `trace_id`, `transaction_id` | Otomatis dari OpenTelemetry |
| **Span (tag)** | `exectime` | Durasi span (ms), diset saat `Finish()` |
| **Span (tag)** | Custom | Via `span.Tag("field", "value")` |
| **Span (tag)** | `error.scope`, `error.message` | Diset saat `span.Error()` |
| **Span (tag)** | `span.status`, `warning.scope`, `warning.message` | Diset saat `span.Warning()` (`span.status=Warning`) |
| **Span (status)** | ERROR | Diset otomatis saat `span.Error()` (span muncul merah di Jaeger) |
| **Event log** | `timestamp` | Waktu log, format `DD/Mon/YYYY HHMMSS.nanosecond` (Asia/Jakarta) |
| **Event log** | `level_id` | `Info`, `Error`, `Warning`, `Debug` |
| **Event log** | `message`, `scope` | Isi pesan dan scope/label |
| **Event log** | `interaction_name`, `interaction_type_name` | Nama dan tipe interaksi |

## Arsitektur (Clean Architecture)

```
tracing/              → API publik (Init, GetTracing, GetTracingForService, NewInteractionName, Shutdown)
domain/               → Entity & interface (LevelId, Span, Tracer, InteractionName/Type)
config/               → Konfigurasi tracer
infrastructure/jaeger/ → Implementasi Jaeger (OTLP)
```

- **Domain**: bebas dependency luar; hanya tipe dan interface.
- **Infrastructure**: implementasi konkret dengan OpenTelemetry + OTLP (Jaeger).

## Instalasi

```bash
go get github.com/funxdofficial/golang-module-jaeger
```

## Inisialisasi (sekali di main)

**Satu service (satu process = satu nama di Jaeger):**

```go
import (
    "context"

    "github.com/funxdofficial/golang-module-jaeger/config"
    "github.com/funxdofficial/golang-module-jaeger/tracing"
)

func main() {
    cfg := config.Config{
        ServiceName: "my-service",
        Endpoint:    "http://localhost:4318", // OTLP HTTP
        ServerIP:    "",                       // kosong = auto-detect
        Insecure:    true,
        SampleRatio: 0.1,                      // ~10% request di-trace; 1.0 untuk debug
    }
    if err := tracing.Init(cfg); err != nil {
        // opsional: tetap jalan dengan no-op tracer
    }
    defer func() {
        _ = tracing.Shutdown(context.Background())
    }()
    // ...
}
```

**Beberapa service (satu process, tampil sebagai banyak service di Jaeger):**

Dari env (comma-separated):

```go
cfg := config.Config{
    ServiceNames: config.ParseServiceNames(os.Getenv("JAEGER_SERVICE_NAMES")), // "order-service, payment-service"
    Endpoint:     "http://localhost:4318",
    Insecure:     true,
}
tracing.Init(cfg)

// Di handler order
span := tracing.GetTracingForService("order-service").Operation(ctx, ...)

// Di handler payment
span := tracing.GetTracingForService("payment-service").Operation(ctx, ...)
```

Dari DB (slice):

```go
cfg := config.Config{
    ServiceNames: serviceNamesFromDB, // []string{"order-service", "payment-service"}
    Endpoint:     "http://localhost:4318",
    Insecure:     true,
}
tracing.Init(cfg)
```

- Jika `ServiceName` kosong dan `ServiceNames` tidak kosong, default tracer memakai `ServiceNames[0]`.
- Jika `ServiceNames` diisi, hanya nama yang ada di daftar yang boleh dipakai di `GetTracingForService` (whitelist).

## Penggunaan di tiap function

Pola yang disarankan:

```go
func InquiryAccount(ctx context.Context) {
    span := tracing.GetTracing().Operation(ctx,
        tracing.NewInteractionName("Controller InquiryAccount"),
        tracing.NewInteractionTypeName("controllers"))
    defer span.Finish()

    span.Tag("user_id", "123")           // tag kustom (muncul di Jaeger)
    span.Tag("request_id", "abc-xyz")

    span.Info("Request", "message")
    span.Debug("Payload", "detail payload")
    span.Warning("Fallback", "using default")
    span.Error("DB", "connection failed")
}
```

Atau jika di project Anda ada package `libs` yang wrap tracing:

```go
package libs

import (
    "github.com/funxdofficial/golang-module-jaeger/domain"
    "github.com/funxdofficial/golang-module-jaeger/tracing"
)

func GetTracing() domain.Tracer {
    return tracing.GetTracing()
}
```

Lalu di controller:

```go
syslog := libs.GetTracing().Operation(ctx,
    tracing.NewInteractionName("Controller InquiryAccount"),
    tracing.NewInteractionTypeName("controllers"))
defer syslog.Finish()
syslog.Tag("field", "value")
syslog.Info("Request", "message")
```

Contoh lengkap: lihat [examples/usage/main.go](examples/usage/main.go).

## Config

| Field | Keterangan |
|-------|------------|
| `ServiceName` | Nama layanan default di Jaeger (untuk `GetTracing`). Kosong + `ServiceNames` ada → pakai `ServiceNames[0]` |
| `ServiceNames` | Daftar nama service (dari DB/env). Untuk `GetTracingForService`; juga dipakai sebagai whitelist |
| `Endpoint` | Endpoint OTLP (misal `http://localhost:4318`). Kosong = no-op tracer |
| `ServerIP` | IP server; kosong = auto-detect atau "-" |
| `Insecure` | `true` untuk HTTP tanpa TLS |
| `SampleRatio` | Rate sampling trace (`0.0`–`1.0`). `0.1` = ~10% request. `<= 0` = semua trace (default SDK `AlwaysOn`) |

- `config.Default()` mengembalikan config default (localhost, `golang-service`, `SampleRatio: 0.1`).
- `config.ParseServiceNames(s string)` memecah string comma-separated jadi `[]string` (trim spasi), berguna untuk env/DB.

### Sampling (`SampleRatio`)

Sampling mengurangi beban CPU/network di server aplikasi. Exporter tetap memakai **batch** (bukan kirim per span).

| Nilai | Perilaku |
|-------|----------|
| `0.1` (default) | ~10% root trace di-record ke Jaeger |
| `1.0` | Semua trace (cocok untuk dev/debug) |
| `<= 0` | Tidak aktifkan ratio sampler → semua trace (`ParentBased(AlwaysOn)`) |

Trace yang **tidak di-sample** tidak mengirim event log/tag ke Jaeger. Operasi jarang dipanggil bisa tidak muncul meski kode sudah benar.

## Status span: Error & Warning

| Method | Event log | Efek di Jaeger |
|--------|-----------|----------------|
| `span.Error(scope, msg)` | `level_id=Error` | Status span **ERROR** + tag `error.scope`, `error.message` |
| `span.Warning(scope, msg)` | `level_id=Warning` | Tag `span.status=Warning`, `warning.scope`, `warning.message` |

OpenTelemetry tidak punya status `Warning`; filter warning di Jaeger via tag `span.status=Warning`. Jika satu span memanggil `Error` dan `Warning`, status tetap **ERROR**.

## Best practice penamaan operation

Nama operation di Jaeger = nilai `NewInteractionName(...)` saat `Operation()`. Gunakan **nama tetap** per handler, jangan masukkan ID/path dinamis:

```go
// ✅ Benar
tracing.NewInteractionName("ActivateKeyHandler.Create")

// ❌ Hindari — membengkak list operation di Jaeger
tracing.NewInteractionName(fmt.Sprintf("Create %s", keyID))
```

Data dinamis → `span.Tag()` atau `span.Info()`, bukan di nama operation.

## Jaeger / OTLP

- Modul ini memakai **OpenTelemetry** dan ekspor via **OTLP (HTTP)**.
- Pastikan ada collector yang menerima OTLP (misal Jaeger 1.35+ dengan OTLP, atau OpenTelemetry Collector) di `Endpoint` yang dikonfigurasi.
- Jika `Endpoint` kosong atau `Init` tidak dipanggil, `GetTracing()` mengembalikan **no-op tracer** (tidak error, tidak kirim data).

## Setup dependency

Jalankan di root project:

```bash
go mod tidy
```

Jika ada error TLS/certificate, jalankan di lingkungan Anda (bukan sandbox) atau sesuaikan `GOPROXY` / root CA.

## Perubahan (Changelog)

Ringkasan perubahan sebelum PR:

### Fitur baru

- **Multi-service dalam satu process** — `config.ServiceNames`, `GetTracingForService(serviceName)`, dan `config.ParseServiceNames()` untuk config dari env/DB. Tracer per service di-cache lazy; `Shutdown()` mematikan semua tracer.
- **Sampling trace** — field `SampleRatio` di config (`ParentBased(TraceIDRatioBased)`) untuk mengurangi overhead server. Default `0.1` via `config.Default()`.
- **Status span Error** — `span.Error()` memanggil `SetStatus(codes.Error)` sehingga span muncul sebagai error di Jaeger UI, plus tag `error.scope` dan `error.message`.
- **Tag span Warning** — `span.Warning()` menambah tag `span.status=Warning`, `warning.scope`, dan `warning.message` (OpenTelemetry tidak punya status Warning native).

### Perubahan API / perilaku

- `Init`: jika `ServiceName` kosong dan `ServiceNames` tidak kosong, default tracer memakai `ServiceNames[0]`.
- `GetTracingForService`: jika `ServiceNames` diisi, hanya nama dalam daftar yang diizinkan (whitelist); selain itu fallback ke tracer default.
- `Shutdown`: shutdown tracer global **dan** semua tracer dari `GetTracingForService`.

### Testing

- `config/config_test.go` — test `Default()` dan `ParseServiceNames()`.
- `tracing/tracing_test.go` — test `Init`, `GetTracing`, `GetTracingForService` (cache, whitelist), dan `Shutdown`.

### Contoh

- `examples/usage/main.go` — contoh single-service + multi-service + `SampleRatio`.

### Breaking changes

Tidak ada breaking change pada signature API publik yang sudah ada (`Init`, `GetTracing`, `Operation`, `span.Info/Error/Warning/Debug`, `Finish`, `Shutdown`). Field config baru (`ServiceNames`, `SampleRatio`) opsional; perilaku lama tetap jalan jika tidak diisi.

## License

MIT
