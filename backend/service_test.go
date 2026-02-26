package main

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeSource struct {
	metadata []MetadataRecord
	details  []DetailsRecord
	err      error

	metadataStart chan struct{}
	metadataGate  <-chan struct{}

	mu                sync.Mutex
	metadataStartOnce sync.Once
	metadataCalls     int
	detailsCalls      int
}

type fakePopularitySource struct {
	records []PopularityRecord
	err     error
}

func (f *fakeSource) LoadMetadata(_ context.Context) ([]MetadataRecord, error) {
	f.mu.Lock()
	start := f.metadataStart
	gate := f.metadataGate
	f.metadataCalls++
	err := f.err
	metadata := append([]MetadataRecord(nil), f.metadata...)
	f.mu.Unlock()

	if start != nil {
		f.metadataStartOnce.Do(func() { close(start) })
	}
	if gate != nil {
		<-gate
	}

	if err != nil {
		return nil, err
	}
	return metadata, nil
}

func (f *fakeSource) LoadDetails(_ context.Context) ([]DetailsRecord, error) {
	f.mu.Lock()
	f.detailsCalls++
	err := f.err
	details := append([]DetailsRecord(nil), f.details...)
	f.mu.Unlock()

	if err != nil {
		return nil, err
	}
	return details, nil
}

func (f *fakeSource) callCounts() (metadataCalls int, detailsCalls int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.metadataCalls, f.detailsCalls
}

func (f *fakeSource) setErr(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.err = err
}

func (f *fakeSource) setMetadataBarrier(start chan struct{}, gate <-chan struct{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.metadataStart = start
	f.metadataGate = gate
	f.metadataStartOnce = sync.Once{}
}

func (f *fakePopularitySource) LoadPopularity(_ context.Context) ([]PopularityRecord, error) {
	if f.err != nil {
		return nil, f.err
	}
	return append([]PopularityRecord(nil), f.records...), nil
}

func TestMergeProducts_BasicAndPriceCalculation(t *testing.T) {
	metadata := []MetadataRecord{
		{ID: "p1", Name: "Phone", BasePrice: 1000, ImageURL: "img", Category: " Smartphones ", Brand: " Apple "},
		{ID: "p2", Name: "Watch", BasePrice: -200, ImageURL: "img2", Category: "Accessories", Brand: "Apple"},
		{ID: "p3", Name: "NoDetails", BasePrice: 100},
	}
	details := []DetailsRecord{
		{ID: "p1", DiscountPercent: 125, Bestseller: true, Colors: []string{"Blue", "blue", "RED"}, Stock: 10, Condition: " Refurbished "},
		{ID: "p2", DiscountPercent: -8, Bestseller: false, Colors: []string{"black"}, Stock: -4, Condition: "Used"},
		{ID: "p999", DiscountPercent: 10, Bestseller: false, Colors: []string{"gray"}},
	}

	got, err := mergeProducts(metadata, details)
	if err != nil {
		t.Fatalf("mergeProducts() unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 merged products, got %d", len(got))
	}

	if got[0].Price != 0 {
		t.Fatalf("expected discounted price clamped to 0, got %v", got[0].Price)
	}
	if strings.Join(got[0].Colors, ",") != "blue,red" {
		t.Fatalf("expected normalized colors [blue red], got %v", got[0].Colors)
	}
	if got[0].DiscountPercent != 100 {
		t.Fatalf("expected discount percent clamped to 100, got %d", got[0].DiscountPercent)
	}
	if got[0].Category != "smartphones" {
		t.Fatalf("expected normalized category smartphones, got %q", got[0].Category)
	}
	if got[0].Brand != "apple" {
		t.Fatalf("expected normalized brand apple, got %q", got[0].Brand)
	}
	if got[0].Condition != "refurbished" {
		t.Fatalf("expected normalized condition refurbished, got %q", got[0].Condition)
	}
	if got[1].Price != 0 {
		t.Fatalf("expected negative base price to clamp to 0, got %v", got[1].Price)
	}
	if got[1].DiscountPercent != 0 {
		t.Fatalf("expected negative discount to clamp to 0, got %d", got[1].DiscountPercent)
	}
	if got[1].Stock != 0 {
		t.Fatalf("expected stock to be clamped to 0, got %d", got[1].Stock)
	}
	if got[1].Condition != "used" {
		t.Fatalf("expected normalized condition used, got %q", got[1].Condition)
	}
}

func TestMergeProducts_UsesStockByColorWhenProvided(t *testing.T) {
	metadata := []MetadataRecord{
		{ID: "p1", Name: "Phone", BasePrice: 200},
	}
	details := []DetailsRecord{
		{
			ID:              "p1",
			DiscountPercent: 10,
			Colors:          []string{"Blue", "Red"},
			Stock:           999,
			StockByColor: map[string]int{
				"blue":  5,
				" red ": -3,
				"green": 2,
			},
		},
	}

	got, err := mergeProducts(metadata, details)
	if err != nil {
		t.Fatalf("mergeProducts() unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected one merged product, got %d", len(got))
	}

	product := got[0]
	if product.Stock != 7 {
		t.Fatalf("expected stock to be summed from stock_by_color (=7), got %d", product.Stock)
	}
	if product.StockByColor["blue"] != 5 {
		t.Fatalf("expected blue stock=5, got %d", product.StockByColor["blue"])
	}
	if product.StockByColor["red"] != 0 {
		t.Fatalf("expected red stock clamped to 0, got %d", product.StockByColor["red"])
	}
	if product.StockByColor["green"] != 2 {
		t.Fatalf("expected green stock=2, got %d", product.StockByColor["green"])
	}
	if strings.Join(product.Colors, ",") != "blue,red,green" {
		t.Fatalf("expected colors [blue red green], got %v", product.Colors)
	}
}

func TestMergeProducts_UsesImageURLsByColorWhenProvided(t *testing.T) {
	metadata := []MetadataRecord{
		{ID: "p1", Name: "Phone", BasePrice: 200},
	}
	details := []DetailsRecord{
		{
			ID:     "p1",
			Colors: []string{"Blue"},
			ImageURLsByColor: map[string]string{
				" blue ": " https://img-blue ",
				"green":  "https://img-green",
				"red":    "   ",
				"":       "https://img-empty-color",
			},
		},
	}

	got, err := mergeProducts(metadata, details)
	if err != nil {
		t.Fatalf("mergeProducts() unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected one merged product, got %d", len(got))
	}

	product := got[0]
	if strings.Join(product.Colors, ",") != "blue,green" {
		t.Fatalf("expected colors [blue green], got %v", product.Colors)
	}
	if len(product.ImageURLsByColor) != 2 {
		t.Fatalf("expected 2 valid image_urls_by_color entries, got %d", len(product.ImageURLsByColor))
	}
	if product.ImageURLsByColor["blue"] != "https://img-blue" {
		t.Fatalf("expected normalized blue image URL, got %q", product.ImageURLsByColor["blue"])
	}
	if product.ImageURLsByColor["green"] != "https://img-green" {
		t.Fatalf("expected normalized green image URL, got %q", product.ImageURLsByColor["green"])
	}
}

func TestMergeProducts_DuplicateIDs(t *testing.T) {
	_, err := mergeProducts(
		[]MetadataRecord{{ID: "p1"}, {ID: "p1"}},
		[]DetailsRecord{{ID: "p1"}},
	)
	if err == nil || !strings.Contains(err.Error(), "duplicate id") {
		t.Fatalf("expected duplicate metadata id error, got %v", err)
	}

	_, err = mergeProducts(
		[]MetadataRecord{{ID: "p1"}},
		[]DetailsRecord{{ID: "p1"}, {ID: "p1"}},
	)
	if err == nil || !strings.Contains(err.Error(), "duplicate id") {
		t.Fatalf("expected duplicate details id error, got %v", err)
	}
}

func TestMergeProducts_EmptyID(t *testing.T) {
	_, err := mergeProducts(
		[]MetadataRecord{{ID: "", Name: "bad"}},
		[]DetailsRecord{{ID: "p1"}},
	)
	if err == nil || !strings.Contains(err.Error(), "metadata contains empty id") {
		t.Fatalf("expected empty metadata id error, got %v", err)
	}

	_, err = mergeProducts(
		[]MetadataRecord{{ID: "p1"}},
		[]DetailsRecord{{ID: ""}},
	)
	if err == nil || !strings.Contains(err.Error(), "details contains empty id") {
		t.Fatalf("expected empty details id error, got %v", err)
	}
}

func TestProductService_QueryProducts_FilterAndPagination(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{
			{ID: "p1", Name: "iPhone 12", BasePrice: 400, Category: "smartphones", Brand: "apple"},
			{ID: "p2", Name: "Galaxy S23", BasePrice: 500, Category: "smartphones", Brand: "samsung"},
			{ID: "p3", Name: "iPhone 13", BasePrice: 700, Category: "smartphones", Brand: "apple"},
		},
		details: []DetailsRecord{
			{ID: "p1", DiscountPercent: 10, Bestseller: true, Colors: []string{"blue"}, Stock: 1, Condition: "refurbished"},
			{ID: "p2", DiscountPercent: 0, Bestseller: false, Colors: []string{"green"}, Stock: 0, Condition: "used"},
			{ID: "p3", DiscountPercent: 20, Bestseller: true, Colors: []string{"red"}, Stock: 1, Condition: "refurbished"},
		},
	}
	service := NewProductService(source, 30*time.Second)

	minPrice := 500.0
	minStock := 1
	limit := 1
	inStock := true
	onSale := true
	response, err := service.QueryProducts(context.Background(), ProductQuery{
		Search:     "iphone",
		Categories: []string{"smartphones"},
		Conditions: []string{"refurbished"},
		Bestseller: boolPtr(true),
		InStock:    &inStock,
		OnSale:     &onSale,
		MinPrice:   &minPrice,
		MinStock:   &minStock,
		Limit:      limit,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("QueryProducts() unexpected error: %v", err)
	}

	if response.Total != 1 {
		t.Fatalf("expected total=1, got %d", response.Total)
	}
	if len(response.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(response.Items))
	}
	if response.Items[0].ID != "p3" {
		t.Fatalf("expected p3, got %s", response.Items[0].ID)
	}
	if response.HasMore {
		t.Fatalf("expected has_more=false")
	}
}

func TestProductService_BrandFilter(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{
			{ID: "p1", Name: "Device A", BasePrice: 400, Brand: "apple"},
			{ID: "p2", Name: "Device B", BasePrice: 500, Brand: "samsung"},
			{ID: "p3", Name: "Device C", BasePrice: 600, Brand: "google"},
		},
		details: []DetailsRecord{
			{ID: "p1", DiscountPercent: 0},
			{ID: "p2", DiscountPercent: 0},
			{ID: "p3", DiscountPercent: 0},
		},
	}
	service := NewProductService(source, 30*time.Second)

	response, err := service.QueryProducts(context.Background(), ProductQuery{
		Brands: []string{"samsung", "google"},
	})
	if err != nil {
		t.Fatalf("QueryProducts() unexpected error: %v", err)
	}

	if response.Total != 2 || len(response.Items) != 2 {
		t.Fatalf("expected two filtered products, got total=%d items=%d", response.Total, len(response.Items))
	}

	for _, item := range response.Items {
		if item.Brand != "samsung" && item.Brand != "google" {
			t.Fatalf("expected only samsung/google brands, got %q", item.Brand)
		}
	}
}

func TestProductService_InStockAndMinStockRespectColorFilter(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{
			{ID: "p1", Name: "Phone A", BasePrice: 500},
		},
		details: []DetailsRecord{
			{
				ID:              "p1",
				DiscountPercent: 0,
				Colors:          []string{"blue", "red"},
				Stock:           50,
				StockByColor: map[string]int{
					"blue": 0,
					"red":  5,
				},
			},
		},
	}
	service := NewProductService(source, 30*time.Second)

	inStockTrue := true
	blueInStockResponse, err := service.QueryProducts(context.Background(), ProductQuery{
		Colors:  []string{"blue"},
		InStock: &inStockTrue,
	})
	if err != nil {
		t.Fatalf("QueryProducts() unexpected error: %v", err)
	}
	if blueInStockResponse.Total != 0 {
		t.Fatalf("expected no products for blue+inStock=true, got total=%d", blueInStockResponse.Total)
	}

	inStockFalse := false
	blueOutOfStockResponse, err := service.QueryProducts(context.Background(), ProductQuery{
		Colors:  []string{"blue"},
		InStock: &inStockFalse,
	})
	if err != nil {
		t.Fatalf("QueryProducts() unexpected error: %v", err)
	}
	if blueOutOfStockResponse.Total != 1 {
		t.Fatalf("expected one product for blue+inStock=false, got total=%d", blueOutOfStockResponse.Total)
	}

	minStock := 1
	blueMinStockResponse, err := service.QueryProducts(context.Background(), ProductQuery{
		Colors:   []string{"blue"},
		MinStock: &minStock,
	})
	if err != nil {
		t.Fatalf("QueryProducts() unexpected error: %v", err)
	}
	if blueMinStockResponse.Total != 0 {
		t.Fatalf("expected no products for blue+minStock=1, got total=%d", blueMinStockResponse.Total)
	}

	redMinStockResponse, err := service.QueryProducts(context.Background(), ProductQuery{
		Colors:   []string{"red"},
		MinStock: &minStock,
	})
	if err != nil {
		t.Fatalf("QueryProducts() unexpected error: %v", err)
	}
	if redMinStockResponse.Total != 1 {
		t.Fatalf("expected one product for red+minStock=1, got total=%d", redMinStockResponse.Total)
	}
}

func TestProductService_OnSaleAndStockFilters(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{
			{ID: "p1", Name: "Phone A", BasePrice: 500},
			{ID: "p2", Name: "Phone B", BasePrice: 500},
			{ID: "p3", Name: "Phone C", BasePrice: 500},
		},
		details: []DetailsRecord{
			{ID: "p1", DiscountPercent: 30, Stock: 2},
			{ID: "p2", DiscountPercent: 0, Stock: 5},
			{ID: "p3", DiscountPercent: 10, Stock: 10},
		},
	}
	service := NewProductService(source, 30*time.Second)

	onSale := false
	minStock := 4

	response, err := service.QueryProducts(context.Background(), ProductQuery{
		OnSale:   &onSale,
		MinStock: &minStock,
	})
	if err != nil {
		t.Fatalf("QueryProducts() unexpected error: %v", err)
	}

	if response.Total != 1 || len(response.Items) != 1 {
		t.Fatalf("expected one filtered product, got total=%d items=%d", response.Total, len(response.Items))
	}
	if response.Items[0].ID != "p2" {
		t.Fatalf("expected p2 to match onSale=false and minStock filter, got %s", response.Items[0].ID)
	}
}

func TestProductService_PriceBoundsInclusive(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{
			{ID: "p1", Name: "Product A", BasePrice: 100},
			{ID: "p2", Name: "Product B", BasePrice: 200},
		},
		details: []DetailsRecord{
			{ID: "p1", DiscountPercent: 0},
			{ID: "p2", DiscountPercent: 0},
		},
	}
	service := NewProductService(source, 30*time.Second)

	minPrice := 100.0
	maxPrice := 100.0
	response, err := service.QueryProducts(context.Background(), ProductQuery{
		MinPrice: &minPrice,
		MaxPrice: &maxPrice,
	})
	if err != nil {
		t.Fatalf("QueryProducts() unexpected error: %v", err)
	}

	if response.Total != 1 || len(response.Items) != 1 {
		t.Fatalf("expected exactly one product at price 100, got total=%d items=%d", response.Total, len(response.Items))
	}
	if response.Items[0].Price != 100 {
		t.Fatalf("expected filtered item price 100, got %v", response.Items[0].Price)
	}
}

func TestProductService_IntegerPricePointIncludesCentPrices(t *testing.T) {
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
	service := NewProductService(source, 30*time.Second)

	minPrice := 100.0
	maxPrice := 100.99
	response, err := service.QueryProducts(context.Background(), ProductQuery{
		MinPrice: &minPrice,
		MaxPrice: &maxPrice,
	})
	if err != nil {
		t.Fatalf("QueryProducts() unexpected error: %v", err)
	}

	if response.Total != 1 || len(response.Items) != 1 {
		t.Fatalf("expected exactly one product within 100..100.99, got total=%d items=%d", response.Total, len(response.Items))
	}
	if response.Items[0].ID != "p1" {
		t.Fatalf("expected p1 at price 100.99, got %s", response.Items[0].ID)
	}
}

func TestProductService_OffsetBeyondTotal_EchoesRequestedOffset(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{{ID: "p1", Name: "Phone", BasePrice: 100}},
		details:  []DetailsRecord{{ID: "p1", DiscountPercent: 0}},
	}
	service := NewProductService(source, 30*time.Second)

	response, err := service.QueryProducts(context.Background(), ProductQuery{
		Limit:  10,
		Offset: 999,
	})
	if err != nil {
		t.Fatalf("QueryProducts() unexpected error: %v", err)
	}

	if response.Total != 1 {
		t.Fatalf("expected total=1, got %d", response.Total)
	}
	if len(response.Items) != 0 {
		t.Fatalf("expected no items, got %d", len(response.Items))
	}
	if response.HasMore {
		t.Fatalf("expected has_more=false")
	}
	if response.Offset != 999 {
		t.Fatalf("expected echoed requested offset=999, got %d", response.Offset)
	}
}

func TestProductService_CachesWithinTTL(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{{ID: "p1", Name: "Phone", BasePrice: 100}},
		details:  []DetailsRecord{{ID: "p1", DiscountPercent: 0}},
	}

	now := time.Date(2026, 2, 24, 19, 0, 0, 0, time.UTC)
	service := NewProductService(source, 30*time.Second)
	service.now = func() time.Time { return now }

	if _, err := service.QueryProducts(context.Background(), ProductQuery{}); err != nil {
		t.Fatalf("first query unexpected error: %v", err)
	}
	if _, err := service.QueryProducts(context.Background(), ProductQuery{}); err != nil {
		t.Fatalf("second query unexpected error: %v", err)
	}

	metadataCalls, detailsCalls := source.callCounts()
	if metadataCalls != 1 || detailsCalls != 1 {
		t.Fatalf("expected exactly one source load in TTL window, metadataCalls=%d detailsCalls=%d", metadataCalls, detailsCalls)
	}
}

func TestProductService_RefreshesAfterTTL(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{{ID: "p1", Name: "Phone", BasePrice: 100}},
		details:  []DetailsRecord{{ID: "p1", DiscountPercent: 0}},
	}

	now := time.Date(2026, 2, 24, 19, 0, 0, 0, time.UTC)
	service := NewProductService(source, 30*time.Second)
	service.now = func() time.Time { return now }

	if _, err := service.QueryProducts(context.Background(), ProductQuery{}); err != nil {
		t.Fatalf("first query unexpected error: %v", err)
	}
	now = now.Add(31 * time.Second)
	if _, err := service.QueryProducts(context.Background(), ProductQuery{}); err != nil {
		t.Fatalf("second query unexpected error: %v", err)
	}

	metadataCalls, detailsCalls := source.callCounts()
	if metadataCalls != 2 || detailsCalls != 2 {
		t.Fatalf("expected cache refresh after TTL, metadataCalls=%d detailsCalls=%d", metadataCalls, detailsCalls)
	}
}

func TestProductService_ServesStaleCacheOnRefreshFailure(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{{ID: "p1", Name: "Phone", BasePrice: 100}},
		details:  []DetailsRecord{{ID: "p1", DiscountPercent: 10}},
	}

	now := time.Date(2026, 2, 24, 19, 0, 0, 0, time.UTC)
	service := NewProductService(source, 30*time.Second)
	service.now = func() time.Time { return now }

	firstResponse, err := service.QueryProducts(context.Background(), ProductQuery{})
	if err != nil {
		t.Fatalf("prime query unexpected error: %v", err)
	}
	if len(firstResponse.Items) != 1 {
		t.Fatalf("expected one item in prime response, got %d", len(firstResponse.Items))
	}

	source.setErr(errors.New("source unavailable"))
	now = now.Add(31 * time.Second)

	staleResponse, err := service.QueryProducts(context.Background(), ProductQuery{})
	if err != nil {
		t.Fatalf("expected stale fallback instead of error, got %v", err)
	}
	if len(staleResponse.Items) != 1 {
		t.Fatalf("expected stale response to include one item, got %d", len(staleResponse.Items))
	}
	if staleResponse.Items[0].ID != firstResponse.Items[0].ID {
		t.Fatalf("expected stale response item id %s, got %s", firstResponse.Items[0].ID, staleResponse.Items[0].ID)
	}
}

func TestProductService_AvoidsCacheStampede(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{{ID: "p1", Name: "Phone", BasePrice: 100}},
		details:  []DetailsRecord{{ID: "p1", DiscountPercent: 0}},
	}

	now := time.Date(2026, 2, 24, 19, 0, 0, 0, time.UTC)
	service := NewProductService(source, 30*time.Second)
	service.now = func() time.Time { return now }

	if _, err := service.QueryProducts(context.Background(), ProductQuery{}); err != nil {
		t.Fatalf("prime query unexpected error: %v", err)
	}

	now = now.Add(31 * time.Second)
	refreshStarted := make(chan struct{})
	releaseRefresh := make(chan struct{})
	source.setMetadataBarrier(refreshStarted, releaseRefresh)

	refreshErr := make(chan error, 1)
	go func() {
		_, err := service.QueryProducts(context.Background(), ProductQuery{})
		refreshErr <- err
	}()

	<-refreshStarted

	const workers = 10
	errCh := make(chan error, workers)
	returned := make(chan struct{}, workers)
	var wg sync.WaitGroup
	var invokedWG sync.WaitGroup
	startWorkers := make(chan struct{})

	invokedWG.Add(workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			invokedWG.Done()
			<-startWorkers
			_, err := service.QueryProducts(context.Background(), ProductQuery{})
			errCh <- err
			returned <- struct{}{}
		}()
	}

	invokedWG.Wait()
	close(startWorkers)

	metadataCalls, detailsCalls := source.callCounts()
	if metadataCalls != 2 || detailsCalls != 1 {
		t.Fatalf("expected one refresh in progress before release, metadataCalls=%d detailsCalls=%d", metadataCalls, detailsCalls)
	}
	select {
	case <-returned:
		t.Fatalf("worker returned before refresh gate was released")
	default:
	}

	close(releaseRefresh)
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent query unexpected error: %v", err)
		}
	}
	if err := <-refreshErr; err != nil {
		t.Fatalf("refresh query unexpected error: %v", err)
	}

	metadataCalls, detailsCalls = source.callCounts()
	if metadataCalls != 2 || detailsCalls != 2 {
		t.Fatalf("expected one refresh load after TTL despite concurrency, metadataCalls=%d detailsCalls=%d", metadataCalls, detailsCalls)
	}
}

func TestProductService_WaitingRequestHonorsCancellation(t *testing.T) {
	started := make(chan struct{})
	releaseMetadata := make(chan struct{})
	source := &fakeSource{
		metadata: []MetadataRecord{{ID: "p1", Name: "Phone", BasePrice: 100}},
		details:  []DetailsRecord{{ID: "p1", DiscountPercent: 0}},
	}
	source.setMetadataBarrier(started, releaseMetadata)
	service := NewProductService(source, 30*time.Second)

	firstErr := make(chan error, 1)
	go func() {
		_, err := service.QueryProducts(context.Background(), ProductQuery{})
		firstErr <- err
	}()

	<-started

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.QueryProducts(ctx, ProductQuery{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected waiting request to honor context cancellation, got %v", err)
	}

	close(releaseMetadata)
	if err := <-firstErr; err != nil {
		t.Fatalf("first query unexpected error: %v", err)
	}
}

func TestProductService_PropagatesSourceError(t *testing.T) {
	source := &fakeSource{err: errors.New("boom")}
	service := NewProductService(source, 30*time.Second)

	_, err := service.QueryProducts(context.Background(), ProductQuery{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "load metadata") {
		t.Fatalf("expected wrapped load metadata error, got %q", err.Error())
	}
}

func TestProductService_EmptyResultsReturnsEmptySlice(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{{ID: "p1", Name: "Phone", BasePrice: 100}},
		details:  []DetailsRecord{{ID: "p1", DiscountPercent: 0, Colors: []string{"blue"}}},
	}
	service := NewProductService(source, 30*time.Second)

	response, err := service.QueryProducts(context.Background(), ProductQuery{
		Colors: []string{"red"},
	})
	if err != nil {
		t.Fatalf("QueryProducts() unexpected error: %v", err)
	}
	if response.Items == nil {
		t.Fatalf("expected empty items slice, got nil")
	}
	if len(response.Items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(response.Items))
	}
}

func TestProductService_AvailableColorsComeFromDataset(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{
			{ID: "p1", Name: "Phone A", BasePrice: 100},
			{ID: "p2", Name: "Phone B", BasePrice: 120},
			{ID: "p3", Name: "Phone C", BasePrice: 140},
		},
		details: []DetailsRecord{
			{ID: "p1", DiscountPercent: 0, Colors: []string{"Blue", "red"}, Stock: 5},
			{ID: "p2", DiscountPercent: 10, Colors: []string{"green", "blue"}, Stock: 3},
			{ID: "p3", DiscountPercent: 10, Colors: []string{"  red  ", ""}, Stock: 1},
		},
	}
	service := NewProductService(source, 30*time.Second)

	response, err := service.QueryProducts(context.Background(), ProductQuery{
		Colors: []string{"red"},
	})
	if err != nil {
		t.Fatalf("QueryProducts() unexpected error: %v", err)
	}

	want := []string{"blue", "green", "red"}
	if strings.Join(response.AvailableColors, ",") != strings.Join(want, ",") {
		t.Fatalf("expected available colors %v, got %v", want, response.AvailableColors)
	}
}

func TestProductService_AvailableColorsExcludeOutOfStockColors(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{
			{ID: "p1", Name: "Phone A", BasePrice: 100},
			{ID: "p2", Name: "Phone B", BasePrice: 120},
		},
		details: []DetailsRecord{
			{
				ID:     "p1",
				Colors: []string{"blue", "red"},
				StockByColor: map[string]int{
					"blue": 4,
					"red":  0,
				},
			},
			{
				ID:     "p2",
				Colors: []string{"red"},
				StockByColor: map[string]int{
					"red": 0,
				},
			},
		},
	}
	service := NewProductService(source, 30*time.Second)

	response, err := service.QueryProducts(context.Background(), ProductQuery{})
	if err != nil {
		t.Fatalf("QueryProducts() unexpected error: %v", err)
	}

	want := []string{"blue"}
	if strings.Join(response.AvailableColors, ",") != strings.Join(want, ",") {
		t.Fatalf("expected available colors %v, got %v", want, response.AvailableColors)
	}
}

func TestProductService_AvailableBrandsComeFromDataset(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{
			{ID: "p1", Name: "Phone A", BasePrice: 100, Brand: " Apple "},
			{ID: "p2", Name: "Phone B", BasePrice: 120, Brand: "samsung"},
			{ID: "p3", Name: "Phone C", BasePrice: 140, Brand: "apple"},
			{ID: "p4", Name: "Phone D", BasePrice: 160, Brand: ""},
		},
		details: []DetailsRecord{
			{ID: "p1", DiscountPercent: 0},
			{ID: "p2", DiscountPercent: 10},
			{ID: "p3", DiscountPercent: 10},
			{ID: "p4", DiscountPercent: 0},
		},
	}
	service := NewProductService(source, 30*time.Second)

	response, err := service.QueryProducts(context.Background(), ProductQuery{
		Search: "phone a",
	})
	if err != nil {
		t.Fatalf("QueryProducts() unexpected error: %v", err)
	}

	want := []string{"apple", "samsung"}
	if strings.Join(response.AvailableBrands, ",") != strings.Join(want, ",") {
		t.Fatalf("expected available brands %v, got %v", want, response.AvailableBrands)
	}
}

func TestProductService_PriceBoundsComeFromDataset(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{
			{ID: "p1", Name: "A", BasePrice: 100},
			{ID: "p2", Name: "B", BasePrice: 200},
			{ID: "p3", Name: "C", BasePrice: 300},
		},
		details: []DetailsRecord{
			{ID: "p1", DiscountPercent: 0},
			{ID: "p2", DiscountPercent: 50},
			{ID: "p3", DiscountPercent: 10},
		},
	}
	service := NewProductService(source, 30*time.Second)

	response, err := service.QueryProducts(context.Background(), ProductQuery{
		Search: "A",
	})
	if err != nil {
		t.Fatalf("QueryProducts() unexpected error: %v", err)
	}

	if response.Total != 1 {
		t.Fatalf("expected filtered total=1, got %d", response.Total)
	}
	if response.PriceMin != 100 {
		t.Fatalf("expected price_min=100 from full dataset, got %v", response.PriceMin)
	}
	if response.PriceMax != 270 {
		t.Fatalf("expected price_max=270 from full dataset, got %v", response.PriceMax)
	}
}

func TestProductService_SortByPopularity(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{
			{ID: "p1", Name: "Alpha", BasePrice: 100},
			{ID: "p2", Name: "Bravo", BasePrice: 100},
			{ID: "p3", Name: "Charlie", BasePrice: 100},
		},
		details: []DetailsRecord{
			{ID: "p1", DiscountPercent: 0},
			{ID: "p2", DiscountPercent: 0},
			{ID: "p3", DiscountPercent: 0},
		},
	}
	popularity := &fakePopularitySource{
		records: []PopularityRecord{
			{ID: "p3", Rank: 1},
			{ID: "p1", Rank: 2},
		},
	}
	service := NewProductService(source, 30*time.Second).WithPopularitySource(popularity)

	response, err := service.QueryProducts(context.Background(), ProductQuery{Sort: "popularity"})
	if err != nil {
		t.Fatalf("QueryProducts() unexpected error: %v", err)
	}
	if len(response.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(response.Items))
	}
	if response.Items[0].ID != "p3" || response.Items[1].ID != "p1" || response.Items[2].ID != "p2" {
		t.Fatalf("expected popularity order [p3 p1 p2], got [%s %s %s]", response.Items[0].ID, response.Items[1].ID, response.Items[2].ID)
	}
	if response.Items[0].PopularityRank != 1 || response.Items[1].PopularityRank != 2 {
		t.Fatalf("expected popularity ranks [1 2 ...], got [%d %d ...]", response.Items[0].PopularityRank, response.Items[1].PopularityRank)
	}
}

func TestProductService_SortByPriceAscAndDesc(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{
			{ID: "p1", Name: "Gamma", BasePrice: 300},
			{ID: "p2", Name: "Alpha", BasePrice: 100},
			{ID: "p3", Name: "Beta", BasePrice: 200},
			{ID: "p4", Name: "Alpha", BasePrice: 200},
			{ID: "p0", Name: "Alpha", BasePrice: 200},
		},
		details: []DetailsRecord{
			{ID: "p1", DiscountPercent: 0},
			{ID: "p2", DiscountPercent: 0},
			{ID: "p3", DiscountPercent: 0},
			{ID: "p4", DiscountPercent: 0},
			{ID: "p0", DiscountPercent: 0},
		},
	}
	service := NewProductService(source, 30*time.Second)

	ascResponse, err := service.QueryProducts(context.Background(), ProductQuery{Sort: "price_asc"})
	if err != nil {
		t.Fatalf("QueryProducts(price_asc) unexpected error: %v", err)
	}
	if len(ascResponse.Items) != 5 {
		t.Fatalf("expected 5 items for price_asc, got %d", len(ascResponse.Items))
	}
	ascIDs := []string{
		ascResponse.Items[0].ID,
		ascResponse.Items[1].ID,
		ascResponse.Items[2].ID,
		ascResponse.Items[3].ID,
		ascResponse.Items[4].ID,
	}
	if strings.Join(ascIDs, ",") != "p2,p0,p4,p3,p1" {
		t.Fatalf("expected price_asc order [p2 p0 p4 p3 p1], got %v", ascIDs)
	}

	descResponse, err := service.QueryProducts(context.Background(), ProductQuery{Sort: "price_desc"})
	if err != nil {
		t.Fatalf("QueryProducts(price_desc) unexpected error: %v", err)
	}
	if len(descResponse.Items) != 5 {
		t.Fatalf("expected 5 items for price_desc, got %d", len(descResponse.Items))
	}
	descIDs := []string{
		descResponse.Items[0].ID,
		descResponse.Items[1].ID,
		descResponse.Items[2].ID,
		descResponse.Items[3].ID,
		descResponse.Items[4].ID,
	}
	if strings.Join(descIDs, ",") != "p1,p0,p4,p3,p2" {
		t.Fatalf("expected price_desc order [p1 p0 p4 p3 p2], got %v", descIDs)
	}
}

func TestProductService_MultiSortPopularityThenPrice(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{
			{ID: "p1", Name: "Gamma", BasePrice: 300},
			{ID: "p2", Name: "Alpha", BasePrice: 100},
			{ID: "p3", Name: "Beta", BasePrice: 200},
		},
		details: []DetailsRecord{
			{ID: "p1", DiscountPercent: 0},
			{ID: "p2", DiscountPercent: 0},
			{ID: "p3", DiscountPercent: 0},
		},
	}
	popularity := &fakePopularitySource{
		records: []PopularityRecord{
			{ID: "p1", Rank: 1},
			{ID: "p2", Rank: 1},
			{ID: "p3", Rank: 2},
		},
	}
	service := NewProductService(source, 30*time.Second).WithPopularitySource(popularity)

	ascResponse, err := service.QueryProducts(context.Background(), ProductQuery{Sort: "popularity,price_asc"})
	if err != nil {
		t.Fatalf("QueryProducts(popularity,price_asc) unexpected error: %v", err)
	}
	ascIDs := []string{ascResponse.Items[0].ID, ascResponse.Items[1].ID, ascResponse.Items[2].ID}
	if strings.Join(ascIDs, ",") != "p2,p1,p3" {
		t.Fatalf("expected popularity+price_asc order [p2 p1 p3], got %v", ascIDs)
	}

	descResponse, err := service.QueryProducts(context.Background(), ProductQuery{Sort: "popularity,price_desc"})
	if err != nil {
		t.Fatalf("QueryProducts(popularity,price_desc) unexpected error: %v", err)
	}
	descIDs := []string{descResponse.Items[0].ID, descResponse.Items[1].ID, descResponse.Items[2].ID}
	if strings.Join(descIDs, ",") != "p1,p2,p3" {
		t.Fatalf("expected popularity+price_desc order [p1 p2 p3], got %v", descIDs)
	}
}

func TestProductService_PopularitySourceFailureDoesNotFailQuery(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{
			{ID: "p1", Name: "Alpha", BasePrice: 100},
		},
		details: []DetailsRecord{
			{ID: "p1", DiscountPercent: 0},
		},
	}
	popularity := &fakePopularitySource{err: errors.New("boom")}
	service := NewProductService(source, 30*time.Second).WithPopularitySource(popularity)

	response, err := service.QueryProducts(context.Background(), ProductQuery{Sort: "popularity"})
	if err != nil {
		t.Fatalf("expected query to succeed even when popularity source fails, got %v", err)
	}
	if response.Total != 1 || len(response.Items) != 1 {
		t.Fatalf("expected one item, got total=%d items=%d", response.Total, len(response.Items))
	}
	if response.Items[0].PopularityRank != 0 {
		t.Fatalf("expected popularity rank to fallback to 0, got %d", response.Items[0].PopularityRank)
	}
}

func TestProductService_PriceBoundsPreservedWhenFilteredResultIsEmpty(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{
			{ID: "p1", Name: "Alpha", BasePrice: 100},
			{ID: "p2", Name: "Bravo", BasePrice: 250},
		},
		details: []DetailsRecord{
			{ID: "p1", DiscountPercent: 10},
			{ID: "p2", DiscountPercent: 0},
		},
	}
	service := NewProductService(source, 30*time.Second)

	response, err := service.QueryProducts(context.Background(), ProductQuery{
		Search: "does-not-exist",
	})
	if err != nil {
		t.Fatalf("QueryProducts() unexpected error: %v", err)
	}

	if response.Total != 0 {
		t.Fatalf("expected filtered total=0, got %d", response.Total)
	}
	if len(response.Items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(response.Items))
	}
	if response.PriceMin != 90 {
		t.Fatalf("expected price_min=90 from full dataset, got %v", response.PriceMin)
	}
	if response.PriceMax != 250 {
		t.Fatalf("expected price_max=250 from full dataset, got %v", response.PriceMax)
	}
}

func TestProductService_PriceBoundsZeroWhenNoMergedProducts(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{
			{ID: "meta-only", Name: "Only Metadata", BasePrice: 123},
		},
		details: []DetailsRecord{
			{ID: "details-only", DiscountPercent: 0},
		},
	}
	service := NewProductService(source, 30*time.Second)

	response, err := service.QueryProducts(context.Background(), ProductQuery{})
	if err != nil {
		t.Fatalf("QueryProducts() unexpected error: %v", err)
	}

	if response.Total != 0 {
		t.Fatalf("expected total=0, got %d", response.Total)
	}
	if response.PriceMin != 0 {
		t.Fatalf("expected price_min=0 when no merged products, got %v", response.PriceMin)
	}
	if response.PriceMax != 0 {
		t.Fatalf("expected price_max=0 when no merged products, got %v", response.PriceMax)
	}
}

func TestDiscountedPriceCents_RoundsAtCentPrecision(t *testing.T) {
	if got := discountedPriceCents(0.01, 50); got != 1 {
		t.Fatalf("expected 1 cent, got %d", got)
	}
	if got := discountedPriceCents(199.99, 25); got != 14999 {
		t.Fatalf("expected 14999 cents, got %d", got)
	}
	if got := discountedPriceCents(414.99, 25); got != 31124 {
		t.Fatalf("expected 31124 cents, got %d", got)
	}
}

func TestProductService_QueryProductsResponseIsImmutableFromCallerMutations(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{
			{ID: "p1", Name: "Alpha", BasePrice: 200, Brand: "Apple"},
		},
		details: []DetailsRecord{
			{
				ID:               "p1",
				DiscountPercent:  10,
				Colors:           []string{"blue"},
				StockByColor:     map[string]int{"blue": 5},
				ImageURLsByColor: map[string]string{"blue": "https://example.com/blue.jpg"},
			},
		},
	}
	service := NewProductService(source, 30*time.Second)

	first, err := service.QueryProducts(context.Background(), ProductQuery{})
	if err != nil {
		t.Fatalf("first QueryProducts() unexpected error: %v", err)
	}
	if len(first.Items) != 1 {
		t.Fatalf("expected one item in first response, got %d", len(first.Items))
	}
	first.Items[0].Colors[0] = "mutated-color"
	first.Items[0].StockByColor["blue"] = 999
	first.Items[0].ImageURLsByColor["blue"] = "https://mutated.invalid/image.jpg"
	first.AvailableColors[0] = "mutated-available-color"
	first.AvailableBrands[0] = "mutated-available-brand"

	second, err := service.QueryProducts(context.Background(), ProductQuery{})
	if err != nil {
		t.Fatalf("second QueryProducts() unexpected error: %v", err)
	}
	if len(second.Items) != 1 {
		t.Fatalf("expected one item in second response, got %d", len(second.Items))
	}
	if second.Items[0].Colors[0] != "blue" {
		t.Fatalf("expected original color to remain blue, got %q", second.Items[0].Colors[0])
	}
	if second.Items[0].StockByColor["blue"] != 5 {
		t.Fatalf("expected original stock_by_color blue=5, got %d", second.Items[0].StockByColor["blue"])
	}
	if second.Items[0].ImageURLsByColor["blue"] != "https://example.com/blue.jpg" {
		t.Fatalf("expected original image_urls_by_color blue URL, got %q", second.Items[0].ImageURLsByColor["blue"])
	}
	if second.AvailableColors[0] != "blue" {
		t.Fatalf("expected available color to remain blue, got %q", second.AvailableColors[0])
	}
	if second.AvailableBrands[0] != "apple" {
		t.Fatalf("expected available brand to remain apple, got %q", second.AvailableBrands[0])
	}
}

func boolPtr(v bool) *bool {
	return &v
}
