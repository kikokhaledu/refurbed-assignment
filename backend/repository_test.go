package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadJSONFile_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := readJSONFile[MetadataRecord](ctx, "ignored.json")
	if err == nil {
		t.Fatalf("expected context cancellation error, got nil")
	}
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestFileProductSource_MissingFile(t *testing.T) {
	source := FileProductSource{
		MetadataPath: filepath.Join("does-not-exist", "metadata.json"),
		DetailsPath:  filepath.Join("does-not-exist", "details.json"),
	}

	_, err := source.LoadMetadata(context.Background())
	if err == nil {
		t.Fatalf("expected metadata missing file error, got nil")
	}
	if !strings.Contains(err.Error(), "read") {
		t.Fatalf("expected read error wrapper, got %q", err.Error())
	}

	_, err = source.LoadDetails(context.Background())
	if err == nil {
		t.Fatalf("expected details missing file error, got nil")
	}
	if !strings.Contains(err.Error(), "read") {
		t.Fatalf("expected read error wrapper, got %q", err.Error())
	}
}

func TestFileProductSource_MalformedJSON(t *testing.T) {
	dir := t.TempDir()
	metadataPath := filepath.Join(dir, "metadata.json")
	detailsPath := filepath.Join(dir, "details.json")

	if err := os.WriteFile(metadataPath, []byte(`[{"id":"p1","name":"Phone","base_price":100}]`), 0o600); err != nil {
		t.Fatalf("failed to write metadata fixture: %v", err)
	}
	if err := os.WriteFile(detailsPath, []byte(`{not-json`), 0o600); err != nil {
		t.Fatalf("failed to write details fixture: %v", err)
	}

	source := FileProductSource{
		MetadataPath: metadataPath,
		DetailsPath:  detailsPath,
	}

	metadata, err := source.LoadMetadata(context.Background())
	if err != nil {
		t.Fatalf("expected metadata to parse, got %v", err)
	}
	if len(metadata) != 1 {
		t.Fatalf("expected 1 metadata record, got %d", len(metadata))
	}

	_, err = source.LoadDetails(context.Background())
	if err == nil {
		t.Fatalf("expected malformed details error, got nil")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Fatalf("expected decode error wrapper, got %q", err.Error())
	}
}

func TestFileProductSource_MissingScalarFieldsDefaultToZeroValues(t *testing.T) {
	dir := t.TempDir()
	metadataPath := filepath.Join(dir, "metadata.json")
	detailsPath := filepath.Join(dir, "details.json")

	if err := os.WriteFile(metadataPath, []byte(`[
		{"id":"p1","name":"Phone","base_price":100}
	]`), 0o600); err != nil {
		t.Fatalf("failed to write metadata fixture: %v", err)
	}
	if err := os.WriteFile(detailsPath, []byte(`[
		{"id":"p1","colors":["blue"],"stock_by_color":{"blue":2}}
	]`), 0o600); err != nil {
		t.Fatalf("failed to write details fixture: %v", err)
	}

	source := FileProductSource{
		MetadataPath: metadataPath,
		DetailsPath:  detailsPath,
	}

	details, err := source.LoadDetails(context.Background())
	if err != nil {
		t.Fatalf("expected details to parse, got %v", err)
	}
	if len(details) != 1 {
		t.Fatalf("expected 1 details record, got %d", len(details))
	}
	if details[0].DiscountPercent != 0 {
		t.Fatalf("expected missing discount_percent to default to 0, got %d", details[0].DiscountPercent)
	}
	if details[0].Stock != 0 {
		t.Fatalf("expected missing stock to default to 0, got %d", details[0].Stock)
	}
	if details[0].Bestseller {
		t.Fatalf("expected missing bestseller to default to false")
	}
}

func TestFileProductSource_NullScalarFieldsDefaultToZeroValues(t *testing.T) {
	dir := t.TempDir()
	metadataPath := filepath.Join(dir, "metadata.json")
	detailsPath := filepath.Join(dir, "details.json")

	if err := os.WriteFile(metadataPath, []byte(`[
		{"id":"p1","name":"Phone","base_price":100}
	]`), 0o600); err != nil {
		t.Fatalf("failed to write metadata fixture: %v", err)
	}
	if err := os.WriteFile(detailsPath, []byte(`[
		{"id":"p1","discount_percent":null,"bestseller":null,"stock":null}
	]`), 0o600); err != nil {
		t.Fatalf("failed to write details fixture: %v", err)
	}

	source := FileProductSource{
		MetadataPath: metadataPath,
		DetailsPath:  detailsPath,
	}

	details, err := source.LoadDetails(context.Background())
	if err != nil {
		t.Fatalf("expected details to parse, got %v", err)
	}
	if len(details) != 1 {
		t.Fatalf("expected 1 details record, got %d", len(details))
	}
	if details[0].DiscountPercent != 0 {
		t.Fatalf("expected null discount_percent to default to 0, got %d", details[0].DiscountPercent)
	}
	if details[0].Stock != 0 {
		t.Fatalf("expected null stock to default to 0, got %d", details[0].Stock)
	}
	if details[0].Bestseller {
		t.Fatalf("expected null bestseller to default to false")
	}
}

func TestFileProductSource_NullMapFieldDecodesToNil(t *testing.T) {
	dir := t.TempDir()
	metadataPath := filepath.Join(dir, "metadata.json")
	detailsPath := filepath.Join(dir, "details.json")

	if err := os.WriteFile(metadataPath, []byte(`[
		{"id":"p1","name":"Phone","base_price":100}
	]`), 0o600); err != nil {
		t.Fatalf("failed to write metadata fixture: %v", err)
	}
	if err := os.WriteFile(detailsPath, []byte(`[
		{"id":"p1","stock_by_color":null}
	]`), 0o600); err != nil {
		t.Fatalf("failed to write details fixture: %v", err)
	}

	source := FileProductSource{
		MetadataPath: metadataPath,
		DetailsPath:  detailsPath,
	}

	details, err := source.LoadDetails(context.Background())
	if err != nil {
		t.Fatalf("expected details to parse, got %v", err)
	}
	if len(details) != 1 {
		t.Fatalf("expected 1 details record, got %d", len(details))
	}
	if details[0].StockByColor != nil {
		t.Fatalf("expected null stock_by_color to decode as nil map")
	}
}

func TestFilePopularitySource_LoadPopularity(t *testing.T) {
	dir := t.TempDir()
	popularityPath := filepath.Join(dir, "popularity.json")

	if err := os.WriteFile(popularityPath, []byte(`[
		{"id":"p2","rank":1},
		{"id":"p1","rank":2}
	]`), 0o600); err != nil {
		t.Fatalf("failed to write popularity fixture: %v", err)
	}

	source := FilePopularitySource{Path: popularityPath}
	records, err := source.LoadPopularity(context.Background())
	if err != nil {
		t.Fatalf("LoadPopularity() unexpected error: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 popularity records, got %d", len(records))
	}
	if records[0].ID != "p2" || records[0].Rank != 1 {
		t.Fatalf("unexpected first popularity record: %+v", records[0])
	}
}
