package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Melikhov-p/go-loyalty-system/internal/contextkeys"
	"github.com/Melikhov-p/go-loyalty-system/internal/models"
	"github.com/Melikhov-p/go-loyalty-system/internal/repository"
	"go.uber.org/zap"
)

func (oh *OrderHandlers) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	user, ok := r.Context().Value(contextkeys.ContextUserKey).(*models.User)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		oh.logger.Error(ErrGettingContextUser.Error())
		return
	}
	if !user.AuthInfo.IsAuthenticated {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	orderNumber, err := io.ReadAll(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	if len(orderNumber) == 0 || !oh.orderService.ValidateOrderNumber(string(orderNumber)) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if err != nil {
		oh.logger.Error("error reading body for create order", zap.Error(err), zap.Int("USER_ID", user.ID))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = oh.orderService.CreateOrder(r.Context(), string(orderNumber), user)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrOrderByUserExist):
			w.WriteHeader(http.StatusOK)
		case errors.Is(err, repository.ErrOrderNumberExist):
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		oh.logger.Error("error creating new order", zap.String("OrderNumber", string(orderNumber)), zap.Error(err))
		return
	}

	oh.logger.Debug("create order", zap.String("NUMBER", string(orderNumber)))
	w.WriteHeader(http.StatusAccepted)
}

func (oh *OrderHandlers) GetOrders(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(contextkeys.ContextUserKey).(*models.User)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		oh.logger.Error(ErrGettingContextUser.Error())
		return
	}
	if !user.AuthInfo.IsAuthenticated {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	orders, err := oh.orderService.GetOrdersByUser(r.Context(), user)
	if err != nil {
		if errors.Is(err, repository.ErrOrdersNotFound) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		oh.logger.Error("error get orders by user", zap.Error(err))
		return
	}

	enc := json.NewEncoder(w)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var ordersResp models.OrdersResponse
	ordersResp.Orders = orders
	if err = enc.Encode(&ordersResp.Orders); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		oh.logger.Error("error encoding response to json", zap.Error(err))
		return
	}
}
