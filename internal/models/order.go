package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Order struct {
	ID         int
	Number     string           `json:"number"`
	Status     string           `json:"status"`
	Accrual    *sql.NullFloat64 `json:"accrual,omitempty"`
	UploadedAt time.Time        `json:"uploaded_at"`
	UserID     int
}

func (o Order) MarshalJSON() ([]byte, error) {
	type Alias Order
	return json.Marshal(&struct {
		Accrual *float64 `json:"accrual,omitempty"`
		*Alias
	}{
		Accrual: o.GetAccrual(),
		Alias:   (*Alias)(&o),
	})
}

// Метод для получения значения Accrual, если оно не NULL
func (o Order) GetAccrual() *float64 {
	if o.Accrual != nil && o.Accrual.Valid {
		return &o.Accrual.Float64
	}
	return nil
}

type WatchedOrder struct {
	ID                 int
	OrderNumber        string
	UserID             int
	AccrualOrderStatus string
}

type AccrualOrderResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}
