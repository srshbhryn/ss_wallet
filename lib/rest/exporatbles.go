package rest

import (
	"context"
	"wallet/lib/rest/internal"
)

type Server interface {
	Run(context.Context) error
	Stop() error
}

var New = internal.New
