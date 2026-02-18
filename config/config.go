package config

// Config konfigurasi untuk Jaeger tracer.
type Config struct {
	// ServiceName nama layanan yang muncul di Jaeger.
	ServiceName string
	// Endpoint endpoint OTLP (misal: http://localhost:4318 untuk HTTP).
	// Kosongkan untuk no-op tracer (tanpa export).
	Endpoint string
	// ServerIP IP server; akan dimasukkan ke setiap log. Kosong = diisi otomatis atau "-".
	ServerIP string
	// Insecure true untuk koneksi HTTP tanpa TLS.
	Insecure bool
}

// Default mengembalikan config default (localhost, nama service default).
func Default() Config {
	return Config{
		ServiceName: "golang-service",
		Endpoint:    "http://localhost:4318",
		Insecure:    true,
	}
}
