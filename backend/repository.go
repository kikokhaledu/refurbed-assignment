package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

type ProductSource interface {
	LoadMetadata(context.Context) ([]MetadataRecord, error)
	LoadDetails(context.Context) ([]DetailsRecord, error)
}

type PopularitySource interface {
	LoadPopularity(context.Context) ([]PopularityRecord, error)
}

type FileProductSource struct {
	MetadataPath string
	DetailsPath  string
}

type FilePopularitySource struct {
	Path string
}

func (s FileProductSource) LoadMetadata(ctx context.Context) ([]MetadataRecord, error) {
	return readJSONFile[MetadataRecord](ctx, s.MetadataPath)
}

func (s FileProductSource) LoadDetails(ctx context.Context) ([]DetailsRecord, error) {
	return readJSONFile[DetailsRecord](ctx, s.DetailsPath)
}

func (s FilePopularitySource) LoadPopularity(ctx context.Context) ([]PopularityRecord, error) {
	return readJSONFile[PopularityRecord](ctx, s.Path)
}

func readJSONFile[T any](ctx context.Context, path string) ([]T, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var records []T
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}

	return records, nil
}
