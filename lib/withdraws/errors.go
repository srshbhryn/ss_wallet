package withdraws

import "errors"

var ErrInsufficientBalance = errors.New("insufficient balance")
var ErrInvalidState = errors.New("cant call this service method for withdraw of this state")
