package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"slices"
	"strings"
	"sync"
	"time"
)

type ProductService struct {
	source           ProductSource
	popularitySource PopularitySource
	ttl              time.Duration
	now              func() time.Time

	mu        sync.Mutex
	cached    *productSnapshot
	expiresAt time.Time
	loading   bool
	loadDone  chan struct{}
}

const staleRetryWindow = 2 * time.Second

type productSnapshot struct {
	products        []Product
	availableColors []string
	availableBrands []string
	priceMin        float64
	priceMax        float64
}

func NewProductService(source ProductSource, ttl time.Duration) *ProductService {
	if ttl <= 0 {
		ttl = DefaultCacheTTLDuration
	}
	return &ProductService{
		source: source,
		ttl:    ttl,
		now:    time.Now,
	}
}

func (s *ProductService) WithPopularitySource(source PopularitySource) *ProductService {
	s.popularitySource = source
	return s
}

func (s *ProductService) QueryProducts(ctx context.Context, query ProductQuery) (ProductListResponse, error) {
	query = sanitizeQuery(query)

	snapshot, err := s.getSnapshot(ctx)
	if err != nil {
		return ProductListResponse{}, err
	}

	filtered := filterProducts(snapshot.products, query)
	sortProducts(filtered, query.Sort)
	total := len(filtered)

	start := query.Offset
	if start > total {
		start = total
	}

	end := start + query.Limit
	if end > total {
		end = total
	}

	page := cloneProducts(filtered[start:end])
	if len(page) == 0 {
		page = []Product{}
	}
	availableColors := cloneStringSlice(snapshot.availableColors)
	availableBrands := cloneStringSlice(snapshot.availableBrands)

	return ProductListResponse{
		Items:           page,
		Total:           total,
		Limit:           query.Limit,
		Offset:          query.Offset,
		HasMore:         end < total,
		AvailableColors: availableColors,
		AvailableBrands: availableBrands,
		PriceMin:        snapshot.priceMin,
		PriceMax:        snapshot.priceMax,
	}, nil
}

func (s *ProductService) getSnapshot(ctx context.Context) (*productSnapshot, error) {
	for {
		now := s.now()

		s.mu.Lock()
		if now.Before(s.expiresAt) && s.cached != nil {
			cached := s.cached
			s.mu.Unlock()
			return cached, nil
		}

		if s.loading {
			loadDone := s.loadDone
			s.mu.Unlock()

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-loadDone:
				continue
			}
		}

		s.loading = true
		s.loadDone = make(chan struct{})
		loadDone := s.loadDone
		s.mu.Unlock()

		snapshot, err := s.loadSnapshot(context.WithoutCancel(ctx))

		var staleFallback *productSnapshot
		s.mu.Lock()
		if err == nil {
			s.cached = snapshot
			s.expiresAt = s.now().Add(s.ttl)
		} else if s.cached != nil {
			staleFallback = s.cached
			retryAfter := staleRetryWindow
			if s.ttl > 0 && s.ttl < retryAfter {
				retryAfter = s.ttl
			}
			if retryAfter <= 0 {
				retryAfter = time.Second
			}
			s.expiresAt = s.now().Add(retryAfter)
		}
		s.loading = false
		close(loadDone)
		s.loadDone = nil
		if err == nil {
			cached := s.cached
			s.mu.Unlock()
			return cached, nil
		}
		s.mu.Unlock()

		if staleFallback != nil {
			log.Printf("products refresh failed, serving stale cache: %v", err)
			return staleFallback, nil
		}

		return nil, err
	}
}

func (s *ProductService) loadSnapshot(ctx context.Context) (*productSnapshot, error) {
	metadata, err := s.source.LoadMetadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("load metadata: %w", err)
	}

	details, err := s.source.LoadDetails(ctx)
	if err != nil {
		return nil, fmt.Errorf("load details: %w", err)
	}

	merged, err := mergeProducts(metadata, details)
	if err != nil {
		return nil, fmt.Errorf("merge products: %w", err)
	}

	applyPopularityRanks(merged, nil)
	if s.popularitySource != nil {
		popularity, popErr := s.popularitySource.LoadPopularity(ctx)
		if popErr != nil {
			log.Printf("popularity source load failed, continuing without popularity sort data: %v", popErr)
			return buildProductSnapshot(merged), nil
		}
		rankings, rankErr := normalizePopularityRankings(popularity)
		if rankErr != nil {
			log.Printf("popularity source data invalid, continuing without popularity sort data: %v", rankErr)
			return buildProductSnapshot(merged), nil
		}
		applyPopularityRanks(merged, rankings)
	}

	return buildProductSnapshot(merged), nil
}

func buildProductSnapshot(products []Product) *productSnapshot {
	availableColors := listAvailableColors(products)
	availableBrands := listAvailableBrands(products)
	priceMin, priceMax := listAvailablePriceBounds(products)

	return &productSnapshot{
		products:        products,
		availableColors: availableColors,
		availableBrands: availableBrands,
		priceMin:        priceMin,
		priceMax:        priceMax,
	}
}

func cloneStringSlice(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

func cloneStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func cloneIntMap(values map[string]int) map[string]int {
	if values == nil {
		return nil
	}
	cloned := make(map[string]int, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func cloneProducts(products []Product) []Product {
	if len(products) == 0 {
		return nil
	}
	cloned := make([]Product, len(products))
	copy(cloned, products)
	for i := range cloned {
		cloned[i].Colors = cloneStringSlice(products[i].Colors)
		cloned[i].ImageURLsByColor = cloneStringMap(products[i].ImageURLsByColor)
		cloned[i].StockByColor = cloneIntMap(products[i].StockByColor)
	}
	return cloned
}

func sanitizeQuery(query ProductQuery) ProductQuery {
	if query.Limit <= 0 {
		query.Limit = defaultLimit
	}
	if query.Limit > maxLimit {
		query.Limit = maxLimit
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	query.Search = strings.TrimSpace(query.Search)
	return query
}

func mergeProducts(metadata []MetadataRecord, details []DetailsRecord) ([]Product, error) {
	detailsByID := make(map[string]DetailsRecord, len(details))
	for _, record := range details {
		id := strings.TrimSpace(record.ID)
		if id == "" {
			return nil, fmt.Errorf("details contains empty id")
		}
		if _, exists := detailsByID[id]; exists {
			return nil, fmt.Errorf("details contains duplicate id %q", id)
		}
		detailsByID[id] = record
	}

	seenMetadataIDs := make(map[string]struct{}, len(metadata))
	products := make([]Product, 0, len(metadata))

	for _, meta := range metadata {
		id := strings.TrimSpace(meta.ID)
		if id == "" {
			return nil, fmt.Errorf("metadata contains empty id")
		}
		if _, exists := seenMetadataIDs[id]; exists {
			return nil, fmt.Errorf("metadata contains duplicate id %q", id)
		}
		seenMetadataIDs[id] = struct{}{}

		detail, ok := detailsByID[id]
		if !ok {
			continue
		}

		baseColors := normalizeColors(detail.Colors)
		stockByColor := normalizeStockByColor(detail.StockByColor, baseColors)
		imageURLsByColor := normalizeImageURLsByColor(detail.ImageURLsByColor)
		colors := mergeColorListWithMapKeys(baseColors, stockByColor)
		colors = mergeColorListWithMapKeys(colors, imageURLsByColor)

		stock := max(0, detail.Stock)
		if len(stockByColor) > 0 {
			stock = sumStockByColor(stockByColor)
		}

		products = append(products, Product{
			ID:               id,
			Name:             strings.TrimSpace(meta.Name),
			Price:            discountedPrice(meta.BasePrice, detail.DiscountPercent),
			DiscountPercent:  clampPercent(detail.DiscountPercent),
			Bestseller:       detail.Bestseller,
			Colors:           colors,
			ImageURLsByColor: imageURLsByColor,
			StockByColor:     stockByColor,
			ImageURL:         strings.TrimSpace(meta.ImageURL),
			Stock:            stock,
			Category:         normalizeToken(meta.Category),
			Brand:            normalizeToken(meta.Brand),
			Condition:        normalizeToken(detail.Condition),
		})
	}

	return products, nil
}

func filterProducts(products []Product, query ProductQuery) []Product {
	if len(products) == 0 {
		return nil
	}

	search := strings.ToLower(strings.TrimSpace(query.Search))
	colorFilter := make(map[string]struct{}, len(query.Colors))
	for _, color := range query.Colors {
		normalized := normalizeToken(color)
		if normalized == "" {
			continue
		}
		colorFilter[normalized] = struct{}{}
	}
	categoryFilter := make(map[string]struct{}, len(query.Categories))
	for _, category := range query.Categories {
		normalized := normalizeToken(category)
		if normalized == "" {
			continue
		}
		categoryFilter[normalized] = struct{}{}
	}
	brandFilter := make(map[string]struct{}, len(query.Brands))
	for _, brand := range query.Brands {
		normalized := normalizeToken(brand)
		if normalized == "" {
			continue
		}
		brandFilter[normalized] = struct{}{}
	}
	conditionFilter := make(map[string]struct{}, len(query.Conditions))
	for _, condition := range query.Conditions {
		normalized := normalizeToken(condition)
		if normalized == "" {
			continue
		}
		conditionFilter[normalized] = struct{}{}
	}

	filtered := make([]Product, 0, len(products))

	for _, product := range products {
		if search != "" && !strings.Contains(strings.ToLower(product.Name), search) {
			continue
		}
		if query.Bestseller != nil && product.Bestseller != *query.Bestseller {
			continue
		}
		if query.MinPrice != nil && product.Price < *query.MinPrice {
			continue
		}
		if query.MaxPrice != nil && product.Price > *query.MaxPrice {
			continue
		}
		if len(colorFilter) > 0 && !matchesAnyColor(product.Colors, colorFilter) {
			continue
		}
		if len(categoryFilter) > 0 {
			if _, ok := categoryFilter[normalizeToken(product.Category)]; !ok {
				continue
			}
		}
		if len(brandFilter) > 0 {
			if _, ok := brandFilter[normalizeToken(product.Brand)]; !ok {
				continue
			}
		}
		if len(conditionFilter) > 0 {
			if _, ok := conditionFilter[normalizeToken(product.Condition)]; !ok {
				continue
			}
		}
		effectiveStock := effectiveStockForQuery(product, colorFilter)
		if query.InStock != nil && (effectiveStock > 0) != *query.InStock {
			continue
		}
		if query.OnSale != nil && (product.DiscountPercent > 0) != *query.OnSale {
			continue
		}
		if query.MinStock != nil && effectiveStock < *query.MinStock {
			continue
		}

		filtered = append(filtered, product)
	}

	return filtered
}

func sortProducts(products []Product, sortMode string) {
	if len(products) <= 1 {
		return
	}

	sortModes := parseSortModes(sortMode)
	if len(sortModes) == 0 {
		return
	}

	slices.SortStableFunc(products, func(a, b Product) int {
		for _, mode := range sortModes {
			var cmp int
			switch mode {
			case SortPopularity:
				cmp = compareByPopularity(a, b)
			case SortPriceAsc:
				cmp = compareByPriceThenNameID(a, b, true)
			case SortPriceDesc:
				cmp = compareByPriceThenNameID(a, b, false)
			default:
				cmp = 0
			}
			if cmp != 0 {
				return cmp
			}
		}
		return compareByNameThenID(a, b)
	})
}

func popularityRankOrFallback(rank int) int {
	if rank > 0 {
		return rank
	}
	return maxInt
}

func parseSortModes(rawSort string) []string {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(rawSort)), ",")
	if len(parts) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(parts))
	modes := make([]string, 0, len(parts))
	for _, part := range parts {
		mode := strings.TrimSpace(part)
		if mode == "" {
			continue
		}
		if !isSupportedSortMode(mode) {
			continue
		}
		if _, exists := seen[mode]; exists {
			continue
		}
		seen[mode] = struct{}{}
		modes = append(modes, mode)
	}
	return modes
}

func compareByPopularity(a, b Product) int {
	aRank := popularityRankOrFallback(a.PopularityRank)
	bRank := popularityRankOrFallback(b.PopularityRank)
	if aRank < bRank {
		return -1
	}
	if aRank > bRank {
		return 1
	}
	return 0
}

func compareByPriceThenNameID(a, b Product, ascending bool) int {
	if a.Price < b.Price {
		if ascending {
			return -1
		}
		return 1
	}
	if a.Price > b.Price {
		if ascending {
			return 1
		}
		return -1
	}
	return compareByNameThenID(a, b)
}

func compareByNameThenID(a, b Product) int {
	aName := strings.ToLower(strings.TrimSpace(a.Name))
	bName := strings.ToLower(strings.TrimSpace(b.Name))
	if aName < bName {
		return -1
	}
	if aName > bName {
		return 1
	}
	if a.ID < b.ID {
		return -1
	}
	if a.ID > b.ID {
		return 1
	}
	return 0
}

func matchesAnyColor(productColors []string, filter map[string]struct{}) bool {
	for _, productColor := range productColors {
		if _, ok := filter[normalizeToken(productColor)]; ok {
			return true
		}
	}
	return false
}

func discountedPrice(basePrice float64, discountPercent int) float64 {
	return float64(discountedPriceCents(basePrice, discountPercent)) / 100
}

func discountedPriceCents(basePrice float64, discountPercent int) int64 {
	if basePrice < 0 {
		basePrice = 0
	}
	baseCents := int64(math.Round(basePrice * 100))

	discountPercent = clampPercent(discountPercent)
	factor := int64(100 - discountPercent)
	return (baseCents*factor + 50) / 100
}

func clampPercent(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func normalizeColors(colors []string) []string {
	seen := make(map[string]struct{}, len(colors))
	out := make([]string, 0, len(colors))
	for _, raw := range colors {
		color := normalizeToken(raw)
		if color == "" {
			continue
		}
		if _, ok := seen[color]; ok {
			continue
		}
		seen[color] = struct{}{}
		out = append(out, color)
	}
	return out
}

func normalizeStockByColor(stockByColor map[string]int, colors []string) map[string]int {
	if len(stockByColor) == 0 {
		return nil
	}

	normalized := make(map[string]int, len(stockByColor)+len(colors))
	for _, rawColor := range colors {
		color := normalizeToken(rawColor)
		if color == "" {
			continue
		}
		normalized[color] = 0
	}

	for rawColor, rawStock := range stockByColor {
		color := normalizeToken(rawColor)
		if color == "" {
			continue
		}
		normalized[color] = max(0, rawStock)
	}

	if len(normalized) == 0 {
		return nil
	}

	return normalized
}

func normalizeImageURLsByColor(imageURLsByColor map[string]string) map[string]string {
	if len(imageURLsByColor) == 0 {
		return nil
	}

	normalized := make(map[string]string, len(imageURLsByColor))
	for rawColor, rawURL := range imageURLsByColor {
		color := normalizeToken(rawColor)
		imageURL := strings.TrimSpace(rawURL)
		if color == "" || imageURL == "" {
			continue
		}
		normalized[color] = imageURL
	}

	if len(normalized) == 0 {
		return nil
	}

	return normalized
}

func mergeColorListWithMapKeys[T any](colors []string, keyedValues map[string]T) []string {
	if len(keyedValues) == 0 {
		return colors
	}

	seen := make(map[string]struct{}, len(colors))
	out := make([]string, 0, len(colors)+len(keyedValues))
	for _, color := range colors {
		normalized := normalizeToken(color)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}

	missing := make([]string, 0, len(keyedValues))
	for color := range keyedValues {
		normalized := normalizeToken(color)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		missing = append(missing, normalized)
	}
	slices.Sort(missing)
	out = append(out, missing...)

	return out
}

func sumStockByColor(stockByColor map[string]int) int {
	if len(stockByColor) == 0 {
		return 0
	}

	total := 0
	for _, stock := range stockByColor {
		total += max(0, stock)
	}
	return total
}

func effectiveStockForQuery(product Product, colorFilter map[string]struct{}) int {
	if len(product.StockByColor) == 0 {
		return max(0, product.Stock)
	}

	if len(colorFilter) == 0 {
		return sumStockByColor(product.StockByColor)
	}

	total := 0
	matched := false
	for color, stock := range product.StockByColor {
		normalized := normalizeToken(color)
		if _, ok := colorFilter[normalized]; !ok {
			continue
		}
		matched = true
		total += max(0, stock)
	}

	if matched {
		return total
	}

	return 0
}

func listAvailableColors(products []Product) []string {
	if len(products) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{})
	out := make([]string, 0)

	for _, product := range products {
		for _, color := range product.Colors {
			normalized := normalizeToken(color)
			if normalized == "" {
				continue
			}
			if !isColorInStockForProduct(product, normalized) {
				continue
			}
			if _, ok := seen[normalized]; ok {
				continue
			}
			seen[normalized] = struct{}{}
			out = append(out, normalized)
		}
	}

	if len(out) == 0 {
		return []string{}
	}

	slices.Sort(out)
	return out
}

func listAvailableBrands(products []Product) []string {
	if len(products) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(products))
	out := make([]string, 0, len(products))

	for _, product := range products {
		brand := normalizeToken(product.Brand)
		if brand == "" {
			continue
		}
		if _, ok := seen[brand]; ok {
			continue
		}
		seen[brand] = struct{}{}
		out = append(out, brand)
	}

	if len(out) == 0 {
		return []string{}
	}

	slices.Sort(out)
	return out
}

func listAvailablePriceBounds(products []Product) (minPrice float64, maxPrice float64) {
	if len(products) == 0 {
		return 0, 0
	}

	minPrice = products[0].Price
	maxPrice = products[0].Price
	for _, product := range products[1:] {
		if product.Price < minPrice {
			minPrice = product.Price
		}
		if product.Price > maxPrice {
			maxPrice = product.Price
		}
	}

	return minPrice, maxPrice
}

func normalizePopularityRankings(records []PopularityRecord) (map[string]int, error) {
	if len(records) == 0 {
		return nil, nil
	}

	rankings := make(map[string]int, len(records))
	for _, record := range records {
		id := strings.TrimSpace(record.ID)
		if id == "" {
			return nil, fmt.Errorf("popularity contains empty id")
		}
		if record.Rank <= 0 {
			return nil, fmt.Errorf("popularity rank for %q must be > 0", id)
		}
		if _, exists := rankings[id]; exists {
			return nil, fmt.Errorf("popularity contains duplicate id %q", id)
		}
		rankings[id] = record.Rank
	}

	return rankings, nil
}

func applyPopularityRanks(products []Product, rankings map[string]int) {
	for i := range products {
		products[i].PopularityRank = 0
		if len(rankings) == 0 {
			continue
		}
		if rank, ok := rankings[products[i].ID]; ok {
			products[i].PopularityRank = rank
		}
	}
}

func isColorInStockForProduct(product Product, color string) bool {
	if len(product.StockByColor) == 0 {
		return max(0, product.Stock) > 0
	}

	stock, ok := product.StockByColor[normalizeToken(color)]
	if !ok {
		return false
	}

	return max(0, stock) > 0
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

const maxInt = int(^uint(0) >> 1)
