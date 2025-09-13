package integrations

import (
	"errors"
	"wallet/lib/withdraws/integrations/internal/common"
)

var ErrDuplicatePayout = common.ErrDuplicatePayout
var ErrInvalidConfig = errors.New("invalid client config")
var ErrUnknownClientType = errors.New("unknown client type")
var ErrClientTypeIsNotImplemented = errors.New("client type is not implemented")
