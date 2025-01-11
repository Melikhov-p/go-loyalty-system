package services

import "errors"

var ErrNotEnough error = errors.New("current balance is not enough for withdraw")
