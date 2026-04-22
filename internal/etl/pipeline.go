package etl

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// Pipeline — оркестратор ETL: extract → transform → load
type Pipeline struct {
	extractor   *Extractor
	transformer *Transformer
	loader      *Loader
	log         *zap.Logger
}

func NewPipeline(e *Extractor, t *Transformer, l *Loader, log *zap.Logger) *Pipeline {
	return &Pipeline{
		extractor:   e,
		transformer: t,
		loader:      l,
		log:         log,
	}
}

// Run запускает полный ETL-цикл для всех заказов
func (p *Pipeline) Run(ctx context.Context) error {
	p.log.Info("etl pipeline started")

	// Extract
	records, err := p.extractor.ExtractAllOrders(ctx)
	if err != nil {
		return fmt.Errorf("pipeline extract: %w", err)
	}
	p.log.Info("extracted records", zap.Int("count", len(records)))

	// Transform
	rows := p.transformer.Transform(records)
	p.log.Info("transformed rows", zap.Int("count", len(rows)))

	// Load
	if err := p.loader.Load(ctx, rows); err != nil {
		return fmt.Errorf("pipeline load: %w", err)
	}

	p.log.Info("etl pipeline finished")
	return nil
}
