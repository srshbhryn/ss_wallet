package integrations

import (
	"wallet/lib/withdraws/enums"
	"wallet/lib/withdraws/integrations/internal/dummy"

	"github.com/mitchellh/mapstructure"
)

type DummyConfig = dummy.Config

type BankClient interface {
	Send(Iban string, amount int64, trackID string) (enums.PayoutStatus, error)
	GetStatus(trackID string) (enums.PayoutStatus, error)
}

func NewBankClient(Type enums.BankType, config any) (BankClient, error) {
	switch Type {
	case enums.DUMMY:
		var cfg dummy.Config
		if err := mapstructure.Decode(config, &cfg); err != nil {
			return nil, ErrInvalidConfig
		}
		return dummy.New(cfg), nil
	case enums.SAMANAN:
		return nil, ErrClientTypeIsNotImplemented
	case enums.MELLAT:
		return nil, ErrClientTypeIsNotImplemented
	default:
		return nil, ErrUnknownClientType
	}
}
