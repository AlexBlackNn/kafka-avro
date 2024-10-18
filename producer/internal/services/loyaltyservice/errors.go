package loyaltyservice

import "errors"

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrNegativeBalance = errors.New("balance must be greater than zero")
)
