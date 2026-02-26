package main

const (
	SortPopularity = "popularity"
	SortPriceAsc   = "price_asc"
	SortPriceDesc  = "price_desc"
)

const sortValidationMessage = "invalid sort: must be one of 'popularity', 'price_asc', 'price_desc'"

func isSupportedSortMode(mode string) bool {
	switch mode {
	case SortPopularity, SortPriceAsc, SortPriceDesc:
		return true
	default:
		return false
	}
}
