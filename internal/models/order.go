package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

// GetAccrual Метод для преобразования sql.NullFloat64 в *float64.
func (o *Order) GetAccrual() *float64 {
	if o.Accrual != nil && o.Accrual.Valid {
		return &o.Accrual.Float64
	}
	return nil
}

func (o *Order) MarshalJSON() ([]byte, error) {
	type Alias Order
	byteArr, err := json.Marshal(&struct {
		Accrual *float64 `json:"accrual,omitempty"`
		*Alias
	}{
		Accrual: o.GetAccrual(),
		Alias:   (*Alias)(o),
	})

	if err != nil {
		return []byte{}, fmt.Errorf("error marshal json for order %w", err)
	}

	return byteArr, nil
}

type OrdersResponse struct {
	Orders []*Order `json:"orders"`
}

type WatchedOrder struct {
	ID                 int
	OrderNumber        string
	UserID             int
	AccrualOrderStatus string
	AccrualPoints      float64
}

type AccrualOrderResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}
