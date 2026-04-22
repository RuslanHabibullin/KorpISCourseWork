package etl

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

// Loader сохраняет трансформированные данные
type Loader struct {
	log    *zap.Logger
	outDir string
}

func NewLoader(log *zap.Logger, outDir string) *Loader {
	if outDir == "" {
		outDir = "./reports"
	}
	return &Loader{log: log, outDir: outDir}
}

// Load сохраняет строки отчёта в JSON-файл
func (l *Loader) Load(_ context.Context, rows []*ReportRow) error {
	if err := os.MkdirAll(l.outDir, 0o755); err != nil {
		return fmt.Errorf("loader: mkdir: %w", err)
	}

	filename := fmt.Sprintf("%s/report_%s.json", l.outDir, time.Now().Format("2006-01-02T15-04-05"))
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("loader: create file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rows); err != nil {
		return fmt.Errorf("loader: encode: %w", err)
	}

	l.log.Info("report saved", zap.String("file", filename), zap.Int("rows", len(rows)))
	return nil
}
