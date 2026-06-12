package config

import (
	"reflect"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.ServiceName != "golang-service" {
		t.Errorf("ServiceName = %q, want golang-service", cfg.ServiceName)
	}
	if cfg.Endpoint != "http://localhost:4318" {
		t.Errorf("Endpoint = %q, want http://localhost:4318", cfg.Endpoint)
	}
	if !cfg.Insecure {
		t.Error("Insecure = false, want true")
	}
}

func TestParseServiceNames(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", nil},
		{"single", "order-service", []string{"order-service"}},
		{"multiple", "order-service,payment-service", []string{"order-service", "payment-service"}},
		{"with spaces", "order-service , payment-service ", []string{"order-service", "payment-service"}},
		{"empty parts", "a,,b", []string{"a", "b"}},
		{"only spaces", "  ,  ", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseServiceNames(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseServiceNames(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
