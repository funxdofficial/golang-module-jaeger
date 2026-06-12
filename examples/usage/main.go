// Contoh penggunaan modul tracing: satu service (GetTracing) dan multi-service (GetTracingForService).
package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/funxdofficial/golang-module-jaeger/config"
	"github.com/funxdofficial/golang-module-jaeger/tracing"
)

func main() {
	// Opsi 1: Satu service (default) dengan sampling ratio untuk mengontrol overhead.
	cfg := config.Config{
		ServiceName: "example-service",
		Endpoint:    "http://localhost:4318",
		ServerIP:    "",
		Insecure:    true,
		// Misal: 0.1 = sekitar 10% request yang di-trace.
		// Untuk traffic sangat tinggi bisa turunkan lagi, mis. 0.01.
		SampleRatio: 0.1,
	}

	// Opsi 2: Multi-service (dari env atau DB); uncomment untuk coba
	// cfg.ServiceNames = config.ParseServiceNames(os.Getenv("JAEGER_SERVICE_NAMES")) // "order-service, payment-service"
	// atau: cfg.ServiceNames = []string{"order-service", "payment-service"}

	if err := tracing.Init(cfg); err != nil {
		log.Printf("tracing init (no-op): %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Contoh: tracer default (satu service)
	InquiryAccount(ctx)

	// Contoh: tracer per service name (muncul terpisah di Jaeger bila ServiceNames di-set)
	CreateOrder(ctx)
	ProcessPayment(ctx)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	if err := tracing.Shutdown(ctx); err != nil {
		log.Printf("tracing shutdown: %v", err)
	}
}

// InquiryAccount memakai tracer default (GetTracing).
func InquiryAccount(ctx context.Context) {
	span := tracing.GetTracing().Operation(ctx,
		tracing.NewInteractionName("Controller InquiryAccount"),
		tracing.NewInteractionTypeName("controllers"))
	defer span.Finish()

	span.Info("Request", "inquiry account request received")
	span.Debug("Payload", "checking payload")
	span.Info("Response", "inquiry account success")
}

// CreateOrder memakai GetTracingForService agar muncul sebagai "order-service" di Jaeger.
func CreateOrder(ctx context.Context) {
	t := tracing.GetTracingForService("order-service")
	span := t.Operation(ctx,
		tracing.NewInteractionName("Handler CreateOrder"),
		tracing.NewInteractionTypeName("handlers"))
	defer span.Finish()

	span.Tag("order_id", "ord-001")
	span.Info("Request", "create order")
	span.Info("Response", "order created")
}

// ProcessPayment memakai GetTracingForService agar muncul sebagai "payment-service" di Jaeger.
func ProcessPayment(ctx context.Context) {
	t := tracing.GetTracingForService("payment-service")
	span := t.Operation(ctx,
		tracing.NewInteractionName("Handler ProcessPayment"),
		tracing.NewInteractionTypeName("handlers"))
	defer span.Finish()

	span.Tag("payment_id", "pay-001")
	span.Info("Request", "process payment")
	span.Info("Response", "payment completed")
}
