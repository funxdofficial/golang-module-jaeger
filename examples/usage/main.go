// Contoh penggunaan modul tracing dengan pola yang diminta.
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
	// Inisialisasi tracer sekali di startup
	cfg := config.Config{
		ServiceName: "example-service",
		Endpoint:    "http://localhost:4318",
		ServerIP:    "",
		Insecure:    true,
	}
	if err := tracing.Init(cfg); err != nil {
		log.Printf("tracing init (no-op): %v", err)
		// Tetap jalan; GetTracing() akan pakai no-op
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Contoh: panggil controller/handler
	InquiryAccount(ctx)

	// Graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	if err := tracing.Shutdown(ctx); err != nil {
		log.Printf("tracing shutdown: %v", err)
	}
}

// InquiryAccount contoh function yang memakai tracing seperti requirement.
func InquiryAccount(ctx context.Context) {
	trace := tracing.GetTracing().Operation(ctx,
		tracing.NewInteractionName("Controller InquiryAccount"),
		tracing.NewInteractionTypeType("controllers"))
	defer trace.Finish()

	trace.Info("Request", "inquiry account request received")
	trace.Debug("Payload", "checking payload")
	// ... logic ...
	trace.Info("Response", "inquiry account success")
}
