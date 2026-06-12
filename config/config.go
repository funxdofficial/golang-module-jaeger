package config

import "strings"

// Config konfigurasi untuk Jaeger tracer.
type Config struct {
	// ServiceName nama layanan default yang muncul di Jaeger (untuk GetTracing).
	// Jika kosong dan ServiceNames tidak kosong, dipakai ServiceNames[0].
	ServiceName string
	// ServiceNames daftar nama service (dari DB/env). Dipakai untuk GetTracingForService.
	// Default tracer pakai ServiceName atau ServiceNames[0] bila ServiceName kosong.
	ServiceNames []string
	// Endpoint endpoint OTLP (misal: http://localhost:4318 untuk HTTP).
	// Kosongkan untuk no-op tracer (tanpa export).
	Endpoint string
	// ServerIP IP server; akan dimasukkan ke setiap log. Kosong = diisi otomatis atau "-".
	ServerIP string
	// Insecure true untuk koneksi HTTP tanpa TLS.
	Insecure bool
	// SampleRatio rate sampling trace (0.0–1.0).
	// 0.0 atau nilai <= 0 artinya pakai default SDK (ParentBased(AlwaysOn)).
	// Misal: 0.1 berarti kira-kira 10% request akan di-trace.
	SampleRatio float64
}

// Default mengembalikan config default (localhost, nama service default).
func Default() Config {
	return Config{
		ServiceName: "golang-service",
		Endpoint:    "http://localhost:4318",
		Insecure:    true,
		// Default sampling ratio; bisa di-override di env.
		SampleRatio: 0.1,
	}
}

// ParseServiceNames memecah string comma-separated jadi slice (trim spasi).
// Berguna saat isi dari env atau kolom DB, misal: "order-service, payment-service".
func ParseServiceNames(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if name := strings.TrimSpace(p); name != "" {
			out = append(out, name)
		}
	}
	return out
}
