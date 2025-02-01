package services

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Melikhov-p/go-loyalty-system/internal/config"
	"github.com/Melikhov-p/go-loyalty-system/internal/models"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const defaultRetryAfter = 60

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

func (as *AccrualService) CheckOrdersStatus(
	order *models.WatchedOrder,
) (*models.WatchedOrder, time.Duration, error) {
	r := resty.New()

	url := fmt.Sprintf("%s/api/orders/%s", as.cfg.AccrualAddr, order.OrderNumber)
	var accrualResp models.AccrualOrderResponse

	resp, err := r.R().SetResult(&accrualResp).Get(url)
	if err != nil {
		return nil, 0, fmt.Errorf("error checking order status %w", err)
	}
	if resp.IsError() {
		if resp.StatusCode() == http.StatusTooManyRequests {
			var seconds int
			secondsStr := resp.Header().Get("Retry-After")
			seconds, err = strconv.Atoi(secondsStr)
			if err != nil {
				as.logger.Error("error parse seconds Retry-After to string")
				seconds = defaultRetryAfter
			}

			return nil, time.Second * time.Duration(seconds), ErrRetryAfter
		}
		return nil, 0, fmt.Errorf("response error %v", resp.Status())
	}

	if accrualResp.Status != order.AccrualOrderStatus {
		order.AccrualOrderStatus = accrualResp.Status
		order.AccrualPoints = accrualResp.Accrual
	}

	return order, 0, nil
}
