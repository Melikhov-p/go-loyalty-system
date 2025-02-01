package handlers

import (
	"net/http"
	"testing"

	"github.com/Melikhov-p/go-loyalty-system/internal/auth"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func OrderTestHandlers(t *testing.T) {
	userToken, err := auth.BuildJWTToken(testUserID, cfg.DB.SecretKey, cfg.TokenLifeTime)
	assert.NoError(t, err)

	t.Run("CREATE ORDER", func(t *testing.T) {
		OrderCreate(t, userToken)
	})
	t.Run("GET USER ORDERS", func(t *testing.T) {
		OrdersGet(t, userToken)
	})

	err = delTestUser(db, "login")
	assert.NoError(t, err)
}

func OrderCreate(t *testing.T, userToken string) {
	testCases := []testCase{
		{
			name:         "Happy create",
			body:         testOrderNumber,
			expectedCode: http.StatusAccepted,
			expectedBody: "",
		},
		{
			name:         "Order Already Exist",
			body:         testOrderNumber,
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name:         "Wrong order format",
			body:         "123",
			expectedCode: http.StatusUnprocessableEntity,
			expectedBody: "",
		},
		{
			name:         "Unauthorized",
			body:         testOrderNumber,
			expectedCode: http.StatusUnauthorized,
			expectedBody: "",
		},
	}

	endPoint := `/api/user/orders`

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			r := resty.New().R()
			r.URL = server.URL + endPoint
			r.Method = http.MethodPost

			r.SetHeader("Content-Type", "text/plain")
			r.SetBody(test.body)

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

func OrdersGet(t *testing.T, userToken string) {
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
			expectedCode: http.StatusUnauthorized,
			expectedBody: "",
		},
	}

	endPoint := `/api/user/orders`

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			r := resty.New().R()
			r.URL = server.URL + endPoint

			if test.expectedCode != http.StatusMethodNotAllowed {
				r.Method = http.MethodGet
			} else {
				r.Method = http.MethodPut
			}

			r.SetHeader("Content-Type", "text/plain")
			r.SetBody(test.body)

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
