package repository

import "errors"

var ErrUserIDNotFound error = errors.New("user not found by provided ID")
var ErrUserWithLoginExist error = errors.New("user with provided login already exist")

var ErrOrderNumberExist error = errors.New("order with provided number already exist")
var ErrOrderByUserExist error = errors.New("order with provided number already created by this user")
var ErrOrdersNotFound error = errors.New("orders by user not found")
var ErrOrderNumberNotFound error = errors.New("order with provided number is not found")

var ErrEmptyBalanceHistory error = errors.New("empty history of withdraws")
