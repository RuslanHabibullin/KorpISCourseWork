package etl

import "fmt"

// ReportRow — трансформированная запись для отчёта
type ReportRow struct {
	OrderID     string
	ClientName  string
	Phone       string
	Vehicle     string // "Brand Model (Plate)"
	Status      string
	Complaint   string
	TotalAmount float64
	PaidTotal   float64
	Debt        float64 // TotalAmount - PaidTotal
	IsPaid      bool
}

// Transformer преобразует raw-записи в отчётные строки
type Transformer struct{}

func NewTransformer() *Transformer {
	return &Transformer{}
}

// Transform конвертирует []Record → []ReportRow
func (t *Transformer) Transform(records []*Record) []*ReportRow {
	rows := make([]*ReportRow, 0, len(records))
	for _, r := range records {
		debt := r.TotalAmount - r.PaidTotal
		if debt < 0 {
			debt = 0
		}
		rows = append(rows, &ReportRow{
			OrderID:     r.OrderID,
			ClientName:  r.ClientName,
			Phone:       r.Phone,
			Vehicle:     fmt.Sprintf("%s %s (%s)", r.Brand, r.Model, r.Plate),
			Status:      r.Status,
			Complaint:   r.Complaint,
			TotalAmount: r.TotalAmount,
			PaidTotal:   r.PaidTotal,
			Debt:        debt,
			IsPaid:      debt == 0 && r.TotalAmount > 0,
		})
	}
	return rows
}
