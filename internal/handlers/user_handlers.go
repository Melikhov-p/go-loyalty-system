package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Melikhov-p/go-loyalty-system/internal/contextkeys"
	"github.com/Melikhov-p/go-loyalty-system/internal/models"
	"github.com/Melikhov-p/go-loyalty-system/internal/repository"
	"github.com/Melikhov-p/go-loyalty-system/internal/services"
	"go.uber.org/zap"
)

func (uh *UserHandlers) UserRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	user, ok := r.Context().Value(contextkeys.ContextUserKey).(*models.User)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		uh.logger.Error(ErrGettingContextUser.Error())
		return
	}
	if user.AuthInfo.IsAuthenticated {
		w.WriteHeader(http.StatusConflict)
		return
	}

	dec := json.NewDecoder(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	var req models.UserLogPassRequest
	if err := dec.Decode(&req); err != nil {
		body, _ := io.ReadAll(r.Body)
		uh.logger.Error("error decoding request", zap.Error(err), zap.String("RAW", string(body)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Login == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := uh.userService.AddNewUser(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrUserWithLoginExist) {
			uh.logger.Error(repository.ErrUserWithLoginExist.Error(), zap.String("LOGIN", req.Login))
			w.WriteHeader(http.StatusConflict)
			return
		}
		uh.logger.Error("error register new user", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "Token",
		Value: user.AuthInfo.Token,
	})
	w.WriteHeader(http.StatusOK)
}

func (uh *UserHandlers) UserLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	user, ok := r.Context().Value(contextkeys.ContextUserKey).(*models.User)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		uh.logger.Error(ErrGettingContextUser.Error())
		return
	}
	if user.AuthInfo.IsAuthenticated {
		http.SetCookie(w, &http.Cookie{
			Name:  "Token",
			Value: user.AuthInfo.Token,
		})
		w.WriteHeader(http.StatusOK)
		return
	}

	dec := json.NewDecoder(r.Body)
	defer func() {
		_ = r.Body.Close()
	}()

	var req models.UserLogPassRequest
	if err := dec.Decode(&req); err != nil {
		uh.logger.Error("error decoding request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Login == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := uh.userService.AuthUser(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrIncorrectPass) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		uh.logger.Error("error authenticate user", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "Token",
		Value: user.AuthInfo.Token,
	})
	w.WriteHeader(http.StatusOK)
}
