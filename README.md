# Golang Module Jaeger

Modul tracing Jaeger untuk Go dengan **clean architecture**, mudah dibaca dan dipakai di berbagai layanan.

- **Process/span**: `server_ip` dan `exectime` (durasi span) ada di process/span Jaeger; `trace_id` dan `transaction_id` otomatis di tag.
- **API**: `GetTracing().Operation(ctx, NewInteractionName(...), NewInteractionTypeName(...))` → `span.Tag()`, `span.Info/Error/Warning/Debug`, `defer span.Finish()`.

## Field di Jaeger

| Lokasi | Field | Keterangan |
|--------|--------|-------------|
| **Process** | `server.ip` | IP server (dari config atau auto-detect) |
| **Span (tag)** | `trace_id`, `transaction_id` | Otomatis dari OpenTelemetry |
| **Span (tag)** | `exectime` | Durasi span (ms), diset saat `Finish()` |
| **Span (tag)** | Custom | Via `span.Tag("field", "value")` |
| **Event log** | `timestamp` | Waktu log, format `DD/Mon/YYYY HHMMSS.nanosecond` (Asia/Jakarta) |
| **Event log** | `level_id` | `Info`, `Error`, `Warning`, `Debug` |
| **Event log** | `message`, `scope` | Isi pesan dan scope/label |
| **Event log** | `interaction_name`, `interaction_type_name` | Nama dan tipe interaksi |

## Arsitektur (Clean Architecture)

```
tracing/              → API publik (Init, GetTracing, NewInteractionName, Shutdown)
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

## Penggunaan di tiap function

Pola yang disarankan:

```go
func InquiryAccount(ctx context.Context) {
    syslog := tracing.GetTracing().Operation(ctx,
        tracing.NewInteractionName("Controller InquiryAccount"),
        tracing.NewInteractionTypeName("controllers"))
    defer span.Finish()

    syslog.Tag("user_id", "123")           // tag kustom (muncul di Jaeger)
    syslog.Tag("request_id", "abc-xyz")

    syslog.Info("Request", "message")
    syslog.Debug("Payload", "detail payload")
    syslog.Warning("Fallback", "using default")
    syslog.Error("DB", "connection failed")
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
| `ServiceName` | Nama layanan di Jaeger |
| `Endpoint` | Endpoint OTLP (misal `http://localhost:4318`). Kosong = no-op tracer |
| `ServerIP` | IP server; kosong = auto-detect atau "-" |
| `Insecure` | `true` untuk HTTP tanpa TLS |

`config.Default()` mengembalikan config default (localhost, `golang-service`).

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

## License

MIT
