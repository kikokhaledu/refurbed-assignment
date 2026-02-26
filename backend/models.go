package main

type MetadataRecord struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	BasePrice float64 `json:"base_price"`
	ImageURL  string  `json:"image_url"`
	Category  string  `json:"category"`
	Brand     string  `json:"brand"`
}

type DetailsRecord struct {
	ID               string            `json:"id"`
	DiscountPercent  int               `json:"discount_percent"`
	Bestseller       bool              `json:"bestseller"`
	Colors           []string          `json:"colors"`
	ImageURLsByColor map[string]string `json:"image_urls_by_color,omitempty"`
	Stock            int               `json:"stock"`
	StockByColor     map[string]int    `json:"stock_by_color"`
	Condition        string            `json:"condition"`
}

type PopularityRecord struct {
	ID   string `json:"id"`
	Rank int    `json:"rank"`
}

type Product struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Price            float64           `json:"price"`
	DiscountPercent  int               `json:"discount_percent"`
	Bestseller       bool              `json:"bestseller"`
	Colors           []string          `json:"colors"`
	ImageURLsByColor map[string]string `json:"image_urls_by_color,omitempty"`
	StockByColor     map[string]int    `json:"stock_by_color"`
	ImageURL         string            `json:"image_url"`
	Stock            int               `json:"stock"`
	Category         string            `json:"category"`
	Brand            string            `json:"brand"`
	Condition        string            `json:"condition"`
	PopularityRank   int               `json:"popularity_rank,omitempty"`
}

type ProductListResponse struct {
	Items           []Product `json:"items"`
	Total           int       `json:"total"`
	Limit           int       `json:"limit"`
	Offset          int       `json:"offset"`
	HasMore         bool      `json:"has_more"`
	AvailableColors []string  `json:"available_colors"`
	AvailableBrands []string  `json:"available_brands"`
	PriceMin        float64   `json:"price_min"`
	PriceMax        float64   `json:"price_max"`
}

type errorResponse struct {
	Error string `json:"error"`
}
