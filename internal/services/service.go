package services

import "errors"

var ErrNotEnough error = errors.New("current balance is not enough for withdraw")
var ErrRetryAfter error = errors.New("got 429 Too Many Requests from Accrual service")
