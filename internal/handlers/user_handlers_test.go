package handlers

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func UserTestHandlers(t *testing.T) {
	defer func() {
		err := delTestUser(db, "login")
		assert.NoError(t, err)
	}()

	t.Run("REGISTER", func(t *testing.T) {
		UserRegister(t)
	})
	t.Run("LOGIN", func(t *testing.T) {
		UserLogin(t)
	})
}

func UserRegister(t *testing.T) {
	login := "login"
	password := "password"

	testCases := []testCase{
		{
			name:         "Happy Register",
			body:         fmt.Sprintf(`{"login":"%s", "password":"%s"}`, login, password),
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name:         "Register existing login",
			body:         fmt.Sprintf(`{"login":"%s", "password":"%s"}`, login, password),
			expectedCode: http.StatusConflict,
			expectedBody: "",
		},
		{
			name:         "Bad Request",
			body:         fmt.Sprintf(`{"login":"%s"}`, login),
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
	}

	endPoint := "/api/user/register"

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			r := resty.New().R()
			r.URL = server.URL + endPoint
			r.Method = http.MethodPost

			r.SetHeader("Content-Type", "application/json")
			r.SetBody(test.body)

			resp, err := r.Send()
			assert.NoError(t, err)

			assert.Equal(t, test.expectedCode, resp.StatusCode())
		})
	}
}

func UserLogin(t *testing.T) {
	testCases := []testCase{
		{
			name:         "Happy login",
			body:         fmt.Sprintf(`{"login":"login", "password": "password"}`),
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name:         "Wrong Pass",
			body:         fmt.Sprintf(`{"login":"login", "password": "pasrd"}`),
			expectedCode: http.StatusUnauthorized,
			expectedBody: "",
		},
		{
			name:         "Empty Pass",
			body:         fmt.Sprintf(`{"login":"login"}`),
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
	}

	endPoint := "/api/user/login"

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			r := resty.New().R()
			r.URL = server.URL + endPoint
			r.Method = http.MethodPost

			r.SetHeader("Content-Type", "application/json")
			r.SetBody(test.body)

			resp, err := r.Send()
			assert.NoError(t, err)

			assert.Equal(t, test.expectedCode, resp.StatusCode())
		})
	}
}
