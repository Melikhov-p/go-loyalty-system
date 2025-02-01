package handlers

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Melikhov-p/go-loyalty-system/internal/auth"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func BalanceTestHandlers(t *testing.T) {
	defer func() {
		err := delTestOrder(db)
		assert.NoError(t, err)
	}()

	userToken, err := auth.BuildJWTToken(testUserID, cfg.DB.SecretKey, cfg.TokenLifeTime)
	assert.NoError(t, err)

	t.Run("GET BALANCE", func(t *testing.T) {
		BalanceGet(t, userToken)
	})
	t.Run("BALANCE WITHDRAW", func(t *testing.T) {
		BalanceWithdraw(t, userToken)
	})
	t.Run("BALANCE WITHDRAW HISTORY", func(t *testing.T) {
		BalanceWithdrawHistory(t, userToken)
	})

	err = delTestUser(db, "login")
	assert.NoError(t, err)
}

func BalanceGet(t *testing.T, userToken string) {
	testCases := []testCase{
		{
			name:         "Happy Getting",
			body:         "",
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name:         "Unauthorized",
			body:         "",
			expectedCode: http.StatusUnauthorized,
			expectedBody: "",
		},
		{
			name:         "Method Not Allowed",
			body:         "",
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
		},
	}

	endPoint := `/api/user/balance`

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			r := resty.New().R()
			r.URL = server.URL + endPoint

			r.SetHeader("Content-Type", "text/plain")
			r.SetBody(test.body)

			if test.expectedCode != http.StatusMethodNotAllowed {
				r.Method = http.MethodGet
			} else {
				r.Method = http.MethodPut
			}

			if test.expectedCode != http.StatusUnauthorized {
				r.SetCookie(&http.Cookie{
					Name:  "Token",
					Value: userToken,
				})
			}

			resp, err := r.Send()
			assert.NoError(t, err)

			assert.Equal(t, test.expectedCode, resp.StatusCode())
		})
	}
}

func BalanceWithdraw(t *testing.T, userToken string) {
	testCases := []testCase{
		{
			name:         "happy withdraw",
			body:         fmt.Sprintf(`{"order":"%s", "sum": 1}`, testOrderNumber),
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name:         "Unauthorized",
			body:         fmt.Sprintf(`{"order":"%s", "sum": 1}`, testOrderNumber),
			expectedCode: http.StatusUnauthorized,
			expectedBody: "",
		},
		{
			name:         "Not Enough",
			body:         fmt.Sprintf(`{"order":"%s", "sum": 999999999}`, testOrderNumber),
			expectedCode: http.StatusPaymentRequired,
			expectedBody: "",
		},
		{
			name:         "Wrong Order Number",
			body:         fmt.Sprintf(`{"order":"%s", "sum": 1}`, "123"),
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
	}

	endPoint := `/api/user/balance/withdraw`

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			r := resty.New().R()
			r.URL = server.URL + endPoint

			r.SetHeader("Content-Type", "application/json")
			r.SetBody(test.body)

			if test.expectedCode != http.StatusMethodNotAllowed {
				r.Method = http.MethodPost
			} else {
				r.Method = http.MethodPut
			}

			if test.expectedCode != http.StatusUnauthorized {
				r.SetCookie(&http.Cookie{
					Name:  "Token",
					Value: userToken,
				})
			}

			resp, err := r.Send()
			assert.NoError(t, err)

			assert.Equal(t, test.expectedCode, resp.StatusCode())
		})
	}
}

func BalanceWithdrawHistory(t *testing.T, userToken string) {
	testCases := []testCase{
		{
			name:         "Happy getting",
			body:         "",
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name:         "Unauthorized",
			body:         "",
			expectedCode: http.StatusUnauthorized,
			expectedBody: "",
		},
		{
			name:         "Method Not Allowed",
			body:         "",
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
		},
	}

	endPoint := `/api/user/withdrawals`

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			r := resty.New().R()
			r.URL = server.URL + endPoint

			r.SetHeader("Content-Type", "text/plain")
			r.SetBody(test.body)

			if test.expectedCode != http.StatusMethodNotAllowed {
				r.Method = http.MethodGet
			} else {
				r.Method = http.MethodPut
			}

			if test.expectedCode != http.StatusUnauthorized {
				r.SetCookie(&http.Cookie{
					Name:  "Token",
					Value: userToken,
				})
			}

			resp, err := r.Send()
			assert.NoError(t, err)

			assert.Equal(t, test.expectedCode, resp.StatusCode())
		})
	}
}
