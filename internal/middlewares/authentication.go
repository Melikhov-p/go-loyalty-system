package middlewares

import (
	"context"
	"errors"
	"net/http"

	"github.com/Melikhov-p/go-loyalty-system/internal/models"
	"github.com/Melikhov-p/go-loyalty-system/internal/services"
	"go.uber.org/zap"
)

func (m *Middleware) WithAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user *models.User
		userService := services.NewUserService(m.logger, m.cfg, m.db)

		tokenCookie, err := r.Cookie("Token")
		if err != nil {
			if !errors.Is(err, http.ErrNoCookie) {
				m.logger.Error("error getting token from cookies", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			user = userService.UserRepo.NewEmptyUser()
		} else {
			tokenString := tokenCookie.Value
			user, err = userService.GetUserByToken(r.Context(), tokenString)
			if err != nil {
				m.logger.Error("error getting user by token", zap.Error(err))

				user = userService.UserRepo.NewEmptyUser()
			} else {
				m.logger.Debug("user authenticated",
					zap.Int("USER_ID", user.ID),
					zap.Float64("USER_BALANCE", user.BalanceInfo.Current))
			}
		}

		ctxWithUser := context.WithValue(r.Context(), "user", user)
		handler.ServeHTTP(w, r.WithContext(ctxWithUser))
	})
}
