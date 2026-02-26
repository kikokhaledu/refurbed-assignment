package main

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const (
	defaultLimit = 6
	maxLimit     = 100
)

var integerPricePattern = regexp.MustCompile(`^\d+$`)

var allowedQueryParams = map[string]struct{}{
	"search":     {},
	"color":      {},
	"category":   {},
	"brand":      {},
	"condition":  {},
	"bestseller": {},
	"inStock":    {},
	"onSale":     {},
	"minStock":   {},
	"minPrice":   {},
	"maxPrice":   {},
	"sort":       {},
	"limit":      {},
	"offset":     {},
}

type ProductQuery struct {
	Search     string
	Colors     []string
	Categories []string
	Brands     []string
	Conditions []string
	Sort       string
	Bestseller *bool
	InStock    *bool
	OnSale     *bool
	MinPrice   *float64
	MaxPrice   *float64
	MinStock   *int
	Limit      int
	Offset     int
}

func ParseProductQuery(values url.Values) (ProductQuery, error) {
	if err := validateAllowedQueryParams(values); err != nil {
		return ProductQuery{}, err
	}

	query := ProductQuery{
		Search:     strings.TrimSpace(values.Get("search")),
		Colors:     parseTokenList(values, "color"),
		Categories: parseTokenList(values, "category"),
		Brands:     parseTokenList(values, "brand"),
		Conditions: parseTokenList(values, "condition"),
		Sort:       "",
		Limit:      defaultLimit,
		Offset:     0,
	}

	bestsellerRaw, hasBestseller, err := singletonQueryValue(values, "bestseller")
	if err != nil {
		return ProductQuery{}, err
	}
	if hasBestseller {
		parsed, err := parseBoolStrict(bestsellerRaw)
		if err != nil {
			return ProductQuery{}, fmt.Errorf("invalid bestseller: %w", err)
		}
		query.Bestseller = &parsed
	}

	inStockRaw, hasInStock, err := singletonQueryValue(values, "inStock")
	if err != nil {
		return ProductQuery{}, err
	}
	if hasInStock {
		parsed, err := parseBoolStrict(inStockRaw)
		if err != nil {
			return ProductQuery{}, fmt.Errorf("invalid inStock: %w", err)
		}
		query.InStock = &parsed
	}
	onSaleRaw, hasOnSale, err := singletonQueryValue(values, "onSale")
	if err != nil {
		return ProductQuery{}, err
	}
	if hasOnSale {
		parsed, err := parseBoolStrict(onSaleRaw)
		if err != nil {
			return ProductQuery{}, fmt.Errorf("invalid onSale: %w", err)
		}
		query.OnSale = &parsed
	}

	minPriceRaw, hasMinPrice, err := singletonQueryValue(values, "minPrice")
	if err != nil {
		return ProductQuery{}, err
	}
	if hasMinPrice {
		parsed, err := parseNonNegativePrice(minPriceRaw, false)
		if err != nil {
			return ProductQuery{}, fmt.Errorf("invalid minPrice: %w", err)
		}
		query.MinPrice = &parsed
	}

	maxPriceRaw, hasMaxPrice, err := singletonQueryValue(values, "maxPrice")
	if err != nil {
		return ProductQuery{}, err
	}
	if hasMaxPrice {
		parsed, err := parseNonNegativePrice(maxPriceRaw, true)
		if err != nil {
			return ProductQuery{}, fmt.Errorf("invalid maxPrice: %w", err)
		}
		query.MaxPrice = &parsed
	}

	minStockRaw, hasMinStock, err := singletonQueryValue(values, "minStock")
	if err != nil {
		return ProductQuery{}, err
	}
	if hasMinStock {
		parsed, err := parseNonNegativeInt(minStockRaw)
		if err != nil {
			return ProductQuery{}, fmt.Errorf("invalid minStock: %w", err)
		}
		query.MinStock = &parsed
	}

	if query.MinPrice != nil && query.MaxPrice != nil && *query.MinPrice > *query.MaxPrice {
		return ProductQuery{}, fmt.Errorf("minPrice cannot be greater than maxPrice")
	}

	sortValue, err := parseSortValue(values)
	if err != nil {
		return ProductQuery{}, err
	}
	query.Sort = sortValue

	limitRaw, hasLimit, err := singletonQueryValue(values, "limit")
	if err != nil {
		return ProductQuery{}, err
	}
	if hasLimit {
		parsed, err := strconv.Atoi(limitRaw)
		if err != nil {
			return ProductQuery{}, fmt.Errorf("invalid limit: must be an integer")
		}
		if parsed <= 0 {
			return ProductQuery{}, fmt.Errorf("invalid limit: must be greater than 0")
		}
		if parsed > maxLimit {
			return ProductQuery{}, fmt.Errorf("invalid limit: must be <= %d", maxLimit)
		}
		query.Limit = parsed
	}

	offsetRaw, hasOffset, err := singletonQueryValue(values, "offset")
	if err != nil {
		return ProductQuery{}, err
	}
	if hasOffset {
		parsed, err := strconv.Atoi(offsetRaw)
		if err != nil {
			return ProductQuery{}, fmt.Errorf("invalid offset: must be an integer")
		}
		if parsed < 0 {
			return ProductQuery{}, fmt.Errorf("invalid offset: must be >= 0")
		}
		query.Offset = parsed
	}

	return query, nil
}

func parseTokenList(values url.Values, key string) []string {
	rawValues := values[key]
	if len(rawValues) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(rawValues))
	items := make([]string, 0, len(rawValues))

	for _, raw := range rawValues {
		parts := strings.Split(raw, ",")
		for _, part := range parts {
			item := normalizeToken(part)
			if item == "" {
				continue
			}
			if _, exists := seen[item]; exists {
				continue
			}
			seen[item] = struct{}{}
			items = append(items, item)
		}
	}

	return items
}

func parseSortValue(values url.Values) (string, error) {
	rawSorts := values["sort"]
	if len(rawSorts) == 0 {
		return "", nil
	}

	seen := make(map[string]struct{}, len(rawSorts))
	sorts := make([]string, 0, len(rawSorts))

	for _, rawSort := range rawSorts {
		parts := strings.Split(rawSort, ",")
		for _, part := range parts {
			sortMode := strings.ToLower(strings.TrimSpace(part))
			if sortMode == "" {
				continue
			}
			if !isSupportedSortMode(sortMode) {
				return "", errors.New(sortValidationMessage)
			}
			if _, exists := seen[sortMode]; exists {
				continue
			}
			seen[sortMode] = struct{}{}
			sorts = append(sorts, sortMode)
		}
	}

	if len(sorts) == 0 {
		return "", nil
	}

	if _, hasAsc := seen[SortPriceAsc]; hasAsc {
		if _, hasDesc := seen[SortPriceDesc]; hasDesc {
			return "", fmt.Errorf("invalid sort: price_asc and price_desc cannot be combined")
		}
	}

	return strings.Join(sorts, ","), nil
}

func singletonQueryValue(values url.Values, key string) (string, bool, error) {
	rawValues, exists := values[key]
	if !exists || len(rawValues) == 0 {
		return "", false, nil
	}
	if len(rawValues) > 1 {
		return "", false, fmt.Errorf("multiple %s values are not allowed", key)
	}

	value := strings.TrimSpace(rawValues[0])
	if value == "" {
		return "", false, fmt.Errorf("empty %s value is not allowed", key)
	}

	return value, true, nil
}

func validateAllowedQueryParams(values url.Values) error {
	for key := range values {
		if _, ok := allowedQueryParams[key]; !ok {
			return fmt.Errorf("unsupported query parameter %q", key)
		}
	}
	return nil
}

func parseBoolStrict(raw string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("must be 'true' or 'false'")
	}
}

func parseNonNegativeFloat(raw string) (float64, error) {
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("must be a number")
	}
	if value < 0 {
		return 0, fmt.Errorf("must be >= 0")
	}
	return value, nil
}

func parseNonNegativePrice(raw string, isUpperBound bool) (float64, error) {
	value, err := parseNonNegativeFloat(raw)
	if err != nil {
		return 0, err
	}

	normalized := strings.TrimSpace(raw)
	if isUpperBound && integerPricePattern.MatchString(normalized) {
		// Slider/UI often send integer euros. Expand the upper bound so
		// maxPrice=712 includes prices up to 712.99.
		value += 0.99
	}

	return value, nil
}

func parseNonNegativeInt(raw string) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("must be an integer")
	}
	if value < 0 {
		return 0, fmt.Errorf("must be >= 0")
	}
	return value, nil
}

func normalizeToken(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}
