package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBuildServerHandler_HealthGet(t *testing.T) {
	logBuffer := captureLogOutput(t)
	service := NewProductService(&fakeSource{}, 30*time.Second)
	handler := buildServerHandler(service, "http://localhost:5173")

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("expected custom CORS origin, got %q", got)
	}
	if vary := recorder.Header().Get("Vary"); !strings.Contains(vary, "Origin") {
		t.Fatalf("expected Vary to include Origin, got %q", vary)
	}

	var payload map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to unmarshal health response: %v", err)
	}
	if payload["status"] != "ok" {
		t.Fatalf("expected health status ok, got %q", payload["status"])
	}

	if !strings.Contains(logBuffer.String(), "GET /health") {
		t.Fatalf("expected request log to include GET /health, got %q", logBuffer.String())
	}
}

func TestBuildServerHandler_HealthMethodNotAllowed(t *testing.T) {
	logBuffer := captureLogOutput(t)
	service := NewProductService(&fakeSource{}, 30*time.Second)
	handler := buildServerHandler(service, "*")

	request := httptest.NewRequest(http.MethodPost, "/health", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("expected Allow=GET, got %q", got)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}

	var payload errorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if payload.Error != "method not allowed" {
		t.Fatalf("expected method not allowed error, got %q", payload.Error)
	}

	if !strings.Contains(logBuffer.String(), "POST /health") {
		t.Fatalf("expected request log to include POST /health, got %q", logBuffer.String())
	}
}

func TestBuildServerHandler_ProductsRouteRegistered(t *testing.T) {
	logBuffer := captureLogOutput(t)
	source := &fakeSource{
		metadata: []MetadataRecord{
			{ID: "p1", Name: "Phone", BasePrice: 100, Brand: "apple"},
		},
		details: []DetailsRecord{
			{ID: "p1", DiscountPercent: 0, Colors: []string{"blue"}, Stock: 2},
		},
	}
	service := NewProductService(source, 30*time.Second)
	handler := buildServerHandler(service, "*")

	request := httptest.NewRequest(http.MethodGet, "/products?limit=1&offset=0", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected wildcard CORS origin, got %q", got)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}

	var payload ProductListResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to unmarshal products response: %v", err)
	}
	if payload.Total != 1 || len(payload.Items) != 1 {
		t.Fatalf("expected one product via registered /products route, got total=%d items=%d", payload.Total, len(payload.Items))
	}

	if !strings.Contains(logBuffer.String(), "GET /products") {
		t.Fatalf("expected request log to include GET /products, got %q", logBuffer.String())
	}
}

func captureLogOutput(t *testing.T) *bytes.Buffer {
	t.Helper()

	origWriter := log.Writer()
	origFlags := log.Flags()
	origPrefix := log.Prefix()

	buffer := &bytes.Buffer{}
	log.SetOutput(buffer)
	log.SetFlags(0)
	log.SetPrefix("")

	t.Cleanup(func() {
		log.SetOutput(origWriter)
		log.SetFlags(origFlags)
		log.SetPrefix(origPrefix)
	})

	return buffer
}
