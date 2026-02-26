package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestProductHandler_MethodNotAllowed(t *testing.T) {
	service := NewProductService(&fakeSource{}, 30*time.Second)
	handler := NewProductHandler(service)

	request := httptest.NewRequest(http.MethodPost, "/products", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", recorder.Code)
	}
	if allow := recorder.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("expected Allow=GET, got %q", allow)
	}
}

func TestProductHandler_InvalidQuery(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{{ID: "p1", Name: "Phone", BasePrice: 100}},
		details:  []DetailsRecord{{ID: "p1", DiscountPercent: 0}},
	}
	handler := NewProductHandler(NewProductService(source, 30*time.Second))

	request := httptest.NewRequest(http.MethodGet, "/products?minPrice=abc", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
	var response errorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if !strings.Contains(response.Error, "invalid minPrice") {
		t.Fatalf("expected minPrice validation error, got %q", response.Error)
	}
}

func TestProductHandler_UnsupportedQueryParameter(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{{ID: "p1", Name: "Phone", BasePrice: 100}},
		details:  []DetailsRecord{{ID: "p1", DiscountPercent: 0}},
	}
	handler := NewProductHandler(NewProductService(source, 30*time.Second))

	request := httptest.NewRequest(http.MethodGet, "/products?foo=price", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
	var response errorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if !strings.Contains(response.Error, "unsupported query parameter") {
		t.Fatalf("expected unsupported query parameter error, got %q", response.Error)
	}
}

func TestProductHandler_Success(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{{ID: "p1", Name: "Phone", BasePrice: 100, Brand: "apple"}},
		details: []DetailsRecord{{
			ID:              "p1",
			DiscountPercent: 20,
			Bestseller:      true,
			Colors:          []string{"blue"},
			ImageURLsByColor: map[string]string{
				"blue": "https://example.com/phone-blue.jpg",
			},
			Stock:        5,
			StockByColor: map[string]int{"blue": 5},
		}},
	}
	handler := NewProductHandler(NewProductService(source, 30*time.Second))

	request := httptest.NewRequest(http.MethodGet, "/products?limit=1&offset=0&sort=popularity", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response ProductListResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Total != 1 || len(response.Items) != 1 {
		t.Fatalf("expected one product in response, got total=%d len=%d", response.Total, len(response.Items))
	}
	if response.Items[0].Price != 80 {
		t.Fatalf("expected discounted price 80, got %v", response.Items[0].Price)
	}
	if response.Items[0].Stock != 5 {
		t.Fatalf("expected stock=5, got %d", response.Items[0].Stock)
	}
	if response.Items[0].StockByColor["blue"] != 5 {
		t.Fatalf("expected stock_by_color.blue=5, got %d", response.Items[0].StockByColor["blue"])
	}
	if response.Items[0].ImageURLsByColor["blue"] != "https://example.com/phone-blue.jpg" {
		t.Fatalf("expected image_urls_by_color.blue to be present, got %q", response.Items[0].ImageURLsByColor["blue"])
	}
	if response.PriceMin != 80 || response.PriceMax != 80 {
		t.Fatalf("expected price bounds min=max=80, got min=%v max=%v", response.PriceMin, response.PriceMax)
	}
	if len(response.AvailableColors) != 1 || response.AvailableColors[0] != "blue" {
		t.Fatalf("expected available_colors [blue], got %v", response.AvailableColors)
	}
	if len(response.AvailableBrands) != 1 || response.AvailableBrands[0] != "apple" {
		t.Fatalf("expected available_brands [apple], got %v", response.AvailableBrands)
	}
}

func TestProductHandler_IntegerPointPriceBucketIncludesCentPrices(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{
			{ID: "p1", Name: "Product A", BasePrice: 100.99},
			{ID: "p2", Name: "Product B", BasePrice: 101.49},
		},
		details: []DetailsRecord{
			{ID: "p1", DiscountPercent: 0},
			{ID: "p2", DiscountPercent: 0},
		},
	}
	handler := NewProductHandler(NewProductService(source, 30*time.Second))

	request := httptest.NewRequest(http.MethodGet, "/products?minPrice=100&maxPrice=100", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response ProductListResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Total != 1 || len(response.Items) != 1 {
		t.Fatalf("expected one product in integer point-price bucket, got total=%d len=%d", response.Total, len(response.Items))
	}
	if response.Items[0].ID != "p1" {
		t.Fatalf("expected product p1 in 100..100.99 bucket, got %s", response.Items[0].ID)
	}
}

func TestCORSMiddleware_Options(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/products", NewProductHandler(NewProductService(&fakeSource{}, 30*time.Second)))
	handler := withCORS(mux, "*")

	request := httptest.NewRequest(http.MethodOptions, "/products", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", recorder.Code)
	}
	if recorder.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin=*")
	}
	if vary := recorder.Header().Get("Vary"); vary != "" {
		t.Fatalf("expected no Vary header for wildcard origin, got %q", vary)
	}
}

func TestCORSMiddleware_CustomOriginSetsVary(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/products", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	handler := withCORS(mux, "http://localhost:5173")

	request := httptest.NewRequest(http.MethodGet, "/products", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("expected Access-Control-Allow-Origin to match configured origin, got %q", got)
	}
	if vary := recorder.Header().Get("Vary"); !strings.Contains(vary, "Origin") {
		t.Fatalf("expected Vary to include Origin, got %q", vary)
	}
}

func TestProductHandler_InternalError(t *testing.T) {
	source := &fakeSource{err: errors.New("source unavailable")}
	handler := NewProductHandler(NewProductService(source, 30*time.Second))

	request := httptest.NewRequest(http.MethodGet, "/products", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", recorder.Code)
	}

	var response errorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if response.Error != "failed to load products" {
		t.Fatalf("expected generic server error message, got %q", response.Error)
	}
}
