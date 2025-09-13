package integrations

import (
	"wallet/lib/withdraws/enums"
	"wallet/lib/withdraws/integrations/internal/dummy"
)

type DummyConfig = dummy.Config

type BankClient interface {
	Send(Iban string, amount int64, trackID string) (enums.PayoutStatus, error)
	GetStatus(trackID string) (enums.PayoutStatus, error)
}

func NewBankClient(Type enums.BankType, config any) (BankClient, error) {
	switch Type {
	case enums.DUMMY:
		cfg, ok := config.(DummyConfig)
		if !ok {
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
