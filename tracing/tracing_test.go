package tracing

import (
	"context"
	"testing"

	"github.com/funxdofficial/golang-module-jaeger/config"
)

func TestGetTracing_BeforeInit(t *testing.T) {
	// Before Init: GetTracing returns something that works (no-op), no panic
	tracer := GetTracing()
	if tracer == nil {
		t.Fatal("GetTracing returned nil")
	}
	ctx := context.Background()
	span := tracer.Operation(ctx, NewInteractionName("test"), NewInteractionTypeName("test"))
	span.Finish()
}

func TestInit_ServiceNamesDefault(t *testing.T) {
	cfg := config.Config{
		ServiceNames: []string{"order-service", "payment-service"},
		Endpoint:     "", // no export, avoid network
		Insecure:     true,
	}
	err := Init(cfg)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer func() {
		_ = Shutdown(context.Background())
	}()

	tracer := GetTracing()
	if tracer == nil {
		t.Fatal("GetTracing returned nil after Init")
	}
	ctx := context.Background()
	span := tracer.Operation(ctx, NewInteractionName("test"), NewInteractionTypeName("test"))
	span.Finish()
}

func TestGetTracingForService_BeforeInit(t *testing.T) {
	// Before Init: GetTracingForService returns something that works, no panic
	tracer := GetTracingForService("order-service")
	if tracer == nil {
		t.Fatal("GetTracingForService returned nil")
	}
	ctx := context.Background()
	span := tracer.Operation(ctx, NewInteractionName("test"), NewInteractionTypeName("test"))
	span.Finish()
}

func TestGetTracingForService_AfterInit(t *testing.T) {
	cfg := config.Config{
		ServiceNames: []string{"order-service", "payment-service"},
		Endpoint:     "",
		Insecure:     true,
	}
	err := Init(cfg)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer func() {
		_ = Shutdown(context.Background())
	}()

	ctx := context.Background()

	t1 := GetTracingForService("order-service")
	span1 := t1.Operation(ctx, NewInteractionName("order"), NewInteractionTypeName("handlers"))
	span1.Finish()

	t2 := GetTracingForService("payment-service")
	span2 := t2.Operation(ctx, NewInteractionName("payment"), NewInteractionTypeName("handlers"))
	span2.Finish()

	// Same name returns same cached tracer
	t1b := GetTracingForService("order-service")
	if t1 != t1b {
		t.Error("GetTracingForService should return cached tracer for same name")
	}
}

func TestGetTracingForService_Whitelist(t *testing.T) {
	cfg := config.Config{
		ServiceNames: []string{"allowed-service"},
		Endpoint:     "",
		Insecure:     true,
	}
	err := Init(cfg)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer func() {
		_ = Shutdown(context.Background())
	}()

	// Not in whitelist: should fall back to default tracer (no error, no panic)
	tracer := GetTracingForService("not-in-list")
	ctx := context.Background()
	span := tracer.Operation(ctx, NewInteractionName("test"), NewInteractionTypeName("test"))
	span.Finish()
}

func TestShutdown(t *testing.T) {
	cfg := config.Config{Endpoint: "", Insecure: true}
	_ = Init(cfg)
	err := Shutdown(context.Background())
	if err != nil {
		t.Errorf("Shutdown: %v", err)
	}
	// After Shutdown, GetTracing still returns a working tracer (no-op), no panic
	tracer := GetTracing()
	if tracer == nil {
		t.Fatal("GetTracing returned nil after Shutdown")
	}
	ctx := context.Background()
	span := tracer.Operation(ctx, NewInteractionName("test"), NewInteractionTypeName("test"))
	span.Finish()
}
