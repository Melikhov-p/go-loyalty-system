package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Melikhov-p/go-loyalty-system/internal/models"
	"github.com/Melikhov-p/go-loyalty-system/internal/repository"
	"github.com/Melikhov-p/go-loyalty-system/internal/services"
	"go.uber.org/zap"
)

func (bh *BalanceHandlers) GetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	user := r.Context().Value("user").(*models.User)
	if !user.AuthInfo.IsAuthenticated {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err := bh.balanceService.GetUserBalance(r.Context(), user)
	if err != nil {
		bh.logger.Error("error getting user balance", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)
	if err = enc.Encode(user.BalanceInfo); err != nil {
		bh.logger.Error("error encoding user balance to json", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (bh *BalanceHandlers) RequestWithdraw(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	user := r.Context().Value("user").(*models.User)
	if !user.AuthInfo.IsAuthenticated {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req *models.WithdrawRequest
	dec := json.NewDecoder(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	if err := dec.Decode(&req); err != nil {
		bh.logger.Error("error decoding withdraw request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	order, err := bh.orderService.GetOrderByNumber(r.Context(), req.Order)
	if err != nil {
		if errors.Is(err, repository.ErrOrdersNotFound) {
			bh.logger.Error("order with gibing number not found", zap.Error(err))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		bh.logger.Error("error getting order with giving number", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user.BalanceInfo, err = bh.balanceService.Withdraw(r.Context(), order, user, req.Sum)
	if err != nil {
		if errors.Is(err, services.ErrNotEnough) {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}

		bh.logger.Error("error withdraw balance", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bh.logger.Debug("withdrawn for user", zap.Int("USER_ID", user.ID), zap.Float64("Withdraw", req.Sum))
	w.WriteHeader(http.StatusOK)
}

func (bh *BalanceHandlers) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	user := r.Context().Value("user").(*models.User)
	if !user.AuthInfo.IsAuthenticated {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	balanceHistory, err := bh.balanceService.GetUserWithdrawHistory(r.Context(), user)
	if err != nil {
		if errors.Is(err, repository.ErrEmptyBalanceHistory) {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		bh.logger.Error("error getting balance history", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")

	if err = enc.Encode(balanceHistory); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		bh.logger.Error("error encoding response for balance history", zap.Error(err))
		return
	}
}
