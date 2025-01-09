package services

import (
	"fmt"

	"github.com/Melikhov-p/go-loyalty-system/internal/config"
	"github.com/Melikhov-p/go-loyalty-system/internal/models"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type AccrualService struct {
	logger *zap.Logger
	cfg    *config.Config
}

func NewAccrualService(logger *zap.Logger, cfg *config.Config) *AccrualService {
	return &AccrualService{
		logger: logger,
		cfg:    cfg,
	}
}

func (as *AccrualService) CheckOrdersStatus(orders []*models.WatchedOrder) ([]*models.WatchedOrder, error) {
	r := resty.New()
	var updatedOrders []*models.WatchedOrder

	for _, order := range orders {
		url := fmt.Sprintf("%s/api/orders/%s", as.cfg.AccrualAddr, order.OrderNumber)
		var accrualResp models.AccrualOrderResponse

		resp, err := r.R().SetResult(&accrualResp).Get(url)
		if err != nil {
			return nil, fmt.Errorf("error checking order status %w", err)
		}
		if resp.IsError() {
			return nil, fmt.Errorf("response error %v", resp.Status())
		}

		if accrualResp.Status != order.AccrualOrderStatus {
			order.AccrualOrderStatus = accrualResp.Status
			order.AccrualPoints = accrualResp.Accrual
			updatedOrders = append(updatedOrders, order)
		}
	}

	return updatedOrders, nil
}
