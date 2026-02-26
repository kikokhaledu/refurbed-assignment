package main

import (
	"net/url"
	"strings"
	"testing"
)

func TestParseProductQuery_Defaults(t *testing.T) {
	values := url.Values{}

	query, err := ParseProductQuery(values)
	if err != nil {
		t.Fatalf("ParseProductQuery() unexpected error: %v", err)
	}

	if query.Limit != defaultLimit {
		t.Fatalf("expected default limit %d, got %d", defaultLimit, query.Limit)
	}
	if query.Offset != 0 {
		t.Fatalf("expected default offset 0, got %d", query.Offset)
	}
	if query.Bestseller != nil {
		t.Fatalf("expected bestseller to be nil")
	}
	if query.InStock != nil {
		t.Fatalf("expected inStock to be nil")
	}
	if query.OnSale != nil {
		t.Fatalf("expected onSale to be nil")
	}
	if query.MinStock != nil {
		t.Fatalf("expected minStock to be nil")
	}
	if len(query.Colors) != 0 {
		t.Fatalf("expected no colors, got %v", query.Colors)
	}
	if len(query.Categories) != 0 {
		t.Fatalf("expected no categories, got %v", query.Categories)
	}
	if len(query.Brands) != 0 {
		t.Fatalf("expected no brands, got %v", query.Brands)
	}
	if len(query.Conditions) != 0 {
		t.Fatalf("expected no conditions, got %v", query.Conditions)
	}
	if query.Sort != "" {
		t.Fatalf("expected default sort to be empty, got %q", query.Sort)
	}
}

func TestParseProductQuery_ParsesFilters(t *testing.T) {
	values := url.Values{
		"search":     []string{"  iPhone "},
		"color":      []string{"Blue, red", "green", "blue"},
		"category":   []string{"Smartphones, tablets", "smartphones"},
		"brand":      []string{"Apple, samsung", "apple"},
		"condition":  []string{"Refurbished, used"},
		"bestseller": []string{"true"},
		"inStock":    []string{"true"},
		"onSale":     []string{"true"},
		"minPrice":   []string{"100"},
		"maxPrice":   []string{"700"},
		"minStock":   []string{"3"},
		"sort":       []string{"popularity"},
		"limit":      []string{"8"},
		"offset":     []string{"16"},
	}

	query, err := ParseProductQuery(values)
	if err != nil {
		t.Fatalf("ParseProductQuery() unexpected error: %v", err)
	}

	if query.Search != "iPhone" {
		t.Fatalf("expected search=iPhone, got %q", query.Search)
	}
	if query.Bestseller == nil || !*query.Bestseller {
		t.Fatalf("expected bestseller=true, got %v", query.Bestseller)
	}
	if query.InStock == nil || !*query.InStock {
		t.Fatalf("expected inStock=true, got %v", query.InStock)
	}
	if query.OnSale == nil || !*query.OnSale {
		t.Fatalf("expected onSale=true, got %v", query.OnSale)
	}
	if query.MinPrice == nil || *query.MinPrice != 100 {
		t.Fatalf("expected minPrice=100, got %v", query.MinPrice)
	}
	if query.MaxPrice == nil || *query.MaxPrice != 700.99 {
		t.Fatalf("expected maxPrice=700.99 for integer upper bound, got %v", query.MaxPrice)
	}
	if query.MinStock == nil || *query.MinStock != 3 {
		t.Fatalf("expected minStock=3, got %v", query.MinStock)
	}
	if query.Sort != "popularity" {
		t.Fatalf("expected sort=popularity, got %q", query.Sort)
	}
	if query.Limit != 8 {
		t.Fatalf("expected limit=8, got %d", query.Limit)
	}
	if query.Offset != 16 {
		t.Fatalf("expected offset=16, got %d", query.Offset)
	}
	wantColors := []string{"blue", "red", "green"}
	if strings.Join(query.Colors, ",") != strings.Join(wantColors, ",") {
		t.Fatalf("expected colors=%v, got %v", wantColors, query.Colors)
	}
	wantCategories := []string{"smartphones", "tablets"}
	if strings.Join(query.Categories, ",") != strings.Join(wantCategories, ",") {
		t.Fatalf("expected categories=%v, got %v", wantCategories, query.Categories)
	}
	wantBrands := []string{"apple", "samsung"}
	if strings.Join(query.Brands, ",") != strings.Join(wantBrands, ",") {
		t.Fatalf("expected brands=%v, got %v", wantBrands, query.Brands)
	}
	wantConditions := []string{"refurbished", "used"}
	if strings.Join(query.Conditions, ",") != strings.Join(wantConditions, ",") {
		t.Fatalf("expected conditions=%v, got %v", wantConditions, query.Conditions)
	}
}

func TestParseProductQuery_SortModes(t *testing.T) {
	tests := []struct {
		name     string
		rawSort  []string
		wantSort string
	}{
		{name: "popularity", rawSort: []string{"popularity"}, wantSort: "popularity"},
		{name: "price asc", rawSort: []string{"price_asc"}, wantSort: "price_asc"},
		{name: "price desc", rawSort: []string{"price_desc"}, wantSort: "price_desc"},
		{name: "case insensitive", rawSort: []string{"PRICE_DESC"}, wantSort: "price_desc"},
		{name: "combined repeated params", rawSort: []string{"popularity", "price_asc"}, wantSort: "popularity,price_asc"},
		{name: "combined comma list", rawSort: []string{"popularity,price_desc"}, wantSort: "popularity,price_desc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			query, err := ParseProductQuery(url.Values{
				"sort": tc.rawSort,
			})
			if err != nil {
				t.Fatalf("ParseProductQuery() unexpected error: %v", err)
			}
			if query.Sort != tc.wantSort {
				t.Fatalf("expected sort=%q, got %q", tc.wantSort, query.Sort)
			}
		})
	}
}

func TestParseProductQuery_ConflictingPriceSortsRejected(t *testing.T) {
	_, err := ParseProductQuery(url.Values{
		"sort": []string{"price_asc", "price_desc"},
	})
	if err == nil {
		t.Fatalf("expected conflicting sort error, got nil")
	}
	if !strings.Contains(err.Error(), "cannot be combined") {
		t.Fatalf("expected cannot be combined error, got %q", err.Error())
	}
}

func TestParseProductQuery_InvalidCases(t *testing.T) {
	tests := []struct {
		name    string
		values  url.Values
		wantErr string
	}{
		{
			name:    "invalid bestseller",
			values:  url.Values{"bestseller": []string{"yes"}},
			wantErr: "invalid bestseller",
		},
		{
			name:    "invalid minPrice",
			values:  url.Values{"minPrice": []string{"abc"}},
			wantErr: "invalid minPrice",
		},
		{
			name:    "invalid inStock",
			values:  url.Values{"inStock": []string{"yes"}},
			wantErr: "invalid inStock",
		},
		{
			name:    "invalid onSale",
			values:  url.Values{"onSale": []string{"yes"}},
			wantErr: "invalid onSale",
		},
		{
			name:    "invalid maxPrice",
			values:  url.Values{"maxPrice": []string{"-1"}},
			wantErr: "invalid maxPrice",
		},
		{
			name:    "invalid minStock",
			values:  url.Values{"minStock": []string{"-1"}},
			wantErr: "invalid minStock",
		},
		{
			name:    "min greater than max",
			values:  url.Values{"minPrice": []string{"500"}, "maxPrice": []string{"100"}},
			wantErr: "minPrice cannot be greater than maxPrice",
		},
		{
			name:    "invalid sort",
			values:  url.Values{"sort": []string{"price"}},
			wantErr: "invalid sort",
		},
		{
			name:    "invalid limit",
			values:  url.Values{"limit": []string{"0"}},
			wantErr: "invalid limit",
		},
		{
			name:    "limit exceeds max",
			values:  url.Values{"limit": []string{"101"}},
			wantErr: "invalid limit",
		},
		{
			name:    "invalid offset",
			values:  url.Values{"offset": []string{"-1"}},
			wantErr: "invalid offset",
		},
		{
			name:    "unsupported query parameter",
			values:  url.Values{"foo": []string{"price"}},
			wantErr: "unsupported query parameter",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseProductQuery(tc.values)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func TestParseProductQuery_Boundaries(t *testing.T) {
	values := url.Values{
		"minPrice": []string{"100.50"},
		"maxPrice": []string{"100.50"},
		"limit":    []string{"100"},
	}

	query, err := ParseProductQuery(values)
	if err != nil {
		t.Fatalf("ParseProductQuery() unexpected error: %v", err)
	}
	if query.MinPrice == nil || *query.MinPrice != 100.5 {
		t.Fatalf("expected minPrice=100.5, got %v", query.MinPrice)
	}
	if query.MaxPrice == nil || *query.MaxPrice != 100.5 {
		t.Fatalf("expected maxPrice=100.5, got %v", query.MaxPrice)
	}
	if query.Limit != 100 {
		t.Fatalf("expected limit=100, got %d", query.Limit)
	}
}

func TestParseProductQuery_IntegerMaxPriceExpandsToEuroCeiling(t *testing.T) {
	query, err := ParseProductQuery(url.Values{
		"minPrice": []string{"712"},
		"maxPrice": []string{"712"},
	})
	if err != nil {
		t.Fatalf("ParseProductQuery() unexpected error: %v", err)
	}

	if query.MinPrice == nil || *query.MinPrice != 712 {
		t.Fatalf("expected minPrice=712, got %v", query.MinPrice)
	}
	if query.MaxPrice == nil || *query.MaxPrice != 712.99 {
		t.Fatalf("expected maxPrice=712.99 for integer upper bound, got %v", query.MaxPrice)
	}
}

func TestParseProductQuery_DegenerateColorInput(t *testing.T) {
	values := url.Values{
		"color": []string{",,", " ", "blue,,", "BLUE"},
	}

	query, err := ParseProductQuery(values)
	if err != nil {
		t.Fatalf("ParseProductQuery() unexpected error: %v", err)
	}

	wantColors := []string{"blue"}
	if strings.Join(query.Colors, ",") != strings.Join(wantColors, ",") {
		t.Fatalf("expected colors=%v, got %v", wantColors, query.Colors)
	}

	onlyDegenerate := url.Values{"color": []string{", , ,", " "}}
	query, err = ParseProductQuery(onlyDegenerate)
	if err != nil {
		t.Fatalf("ParseProductQuery() unexpected error: %v", err)
	}
	if len(query.Colors) != 0 {
		t.Fatalf("expected no colors for degenerate input, got %v", query.Colors)
	}

	degenerateCategory := url.Values{"category": []string{", ,", "  "}}
	query, err = ParseProductQuery(degenerateCategory)
	if err != nil {
		t.Fatalf("ParseProductQuery() unexpected error: %v", err)
	}
	if len(query.Categories) != 0 {
		t.Fatalf("expected no categories for degenerate input, got %v", query.Categories)
	}

	degenerateCondition := url.Values{"condition": []string{"", " , "}}
	query, err = ParseProductQuery(degenerateCondition)
	if err != nil {
		t.Fatalf("ParseProductQuery() unexpected error: %v", err)
	}
	if len(query.Conditions) != 0 {
		t.Fatalf("expected no conditions for degenerate input, got %v", query.Conditions)
	}

	degenerateBrand := url.Values{"brand": []string{", ,", "  "}}
	query, err = ParseProductQuery(degenerateBrand)
	if err != nil {
		t.Fatalf("ParseProductQuery() unexpected error: %v", err)
	}
	if len(query.Brands) != 0 {
		t.Fatalf("expected no brands for degenerate input, got %v", query.Brands)
	}
}

func TestParseProductQuery_RejectsRepeatedSingletonParams(t *testing.T) {
	tests := []struct {
		name    string
		values  url.Values
		wantErr string
	}{
		{
			name:    "repeated bestseller",
			values:  url.Values{"bestseller": []string{"true", "false"}},
			wantErr: "multiple bestseller values are not allowed",
		},
		{
			name:    "repeated inStock",
			values:  url.Values{"inStock": []string{"true", "false"}},
			wantErr: "multiple inStock values are not allowed",
		},
		{
			name:    "repeated onSale",
			values:  url.Values{"onSale": []string{"true", "false"}},
			wantErr: "multiple onSale values are not allowed",
		},
		{
			name:    "repeated minPrice",
			values:  url.Values{"minPrice": []string{"100", "200"}},
			wantErr: "multiple minPrice values are not allowed",
		},
		{
			name:    "repeated maxPrice",
			values:  url.Values{"maxPrice": []string{"100", "200"}},
			wantErr: "multiple maxPrice values are not allowed",
		},
		{
			name:    "repeated minStock",
			values:  url.Values{"minStock": []string{"1", "2"}},
			wantErr: "multiple minStock values are not allowed",
		},
		{
			name:    "repeated limit",
			values:  url.Values{"limit": []string{"6", "12"}},
			wantErr: "multiple limit values are not allowed",
		},
		{
			name:    "repeated offset",
			values:  url.Values{"offset": []string{"0", "6"}},
			wantErr: "multiple offset values are not allowed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseProductQuery(tc.values)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func TestParseProductQuery_RejectsEmptySingletonParams(t *testing.T) {
	tests := []struct {
		name    string
		values  url.Values
		wantErr string
	}{
		{
			name:    "empty bestseller",
			values:  url.Values{"bestseller": []string{""}},
			wantErr: "empty bestseller value is not allowed",
		},
		{
			name:    "whitespace inStock",
			values:  url.Values{"inStock": []string{"  "}},
			wantErr: "empty inStock value is not allowed",
		},
		{
			name:    "empty onSale",
			values:  url.Values{"onSale": []string{""}},
			wantErr: "empty onSale value is not allowed",
		},
		{
			name:    "empty minPrice",
			values:  url.Values{"minPrice": []string{""}},
			wantErr: "empty minPrice value is not allowed",
		},
		{
			name:    "empty maxPrice",
			values:  url.Values{"maxPrice": []string{""}},
			wantErr: "empty maxPrice value is not allowed",
		},
		{
			name:    "empty minStock",
			values:  url.Values{"minStock": []string{""}},
			wantErr: "empty minStock value is not allowed",
		},
		{
			name:    "empty limit",
			values:  url.Values{"limit": []string{""}},
			wantErr: "empty limit value is not allowed",
		},
		{
			name:    "empty offset",
			values:  url.Values{"offset": []string{""}},
			wantErr: "empty offset value is not allowed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseProductQuery(tc.values)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}
