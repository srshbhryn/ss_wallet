package dummy

import (
	"math/rand/v2"
	"sync"
	"wallet/lib/withdraws/enums"
	"wallet/lib/withdraws/integrations/internal/common"
)

type Config struct {
	FailureRate float64
}

type dummyClient struct {
	failureRate     float64
	traceIDSet      map[string]struct{}
	traceIDSetMutex sync.RWMutex
}

func New(config Config) *dummyClient {
	return &dummyClient{
		failureRate: config.FailureRate,
		traceIDSet:  make(map[string]struct{}, 256),
	}
}

func (c *dummyClient) Send(Iban string, amount int64, trackID string) (enums.PayoutStatus, error) {
	isDuplicate := func() bool {
		c.traceIDSetMutex.RLock()
		defer c.traceIDSetMutex.RUnlock()
		_, ok := c.traceIDSet[trackID]
		return ok
	}()
	if isDuplicate {
		return enums.PayoutStatus(""), common.ErrDuplicatePayout
	}
	isSuccessful := rand.Float64() > c.failureRate
	if !isSuccessful {
		return enums.FAILED, nil
	}
	c.traceIDSetMutex.Lock()
	defer c.traceIDSetMutex.Unlock()
	c.traceIDSet[trackID] = struct{}{}
	return enums.SUCCESS, nil

}
func (c *dummyClient) GetStatus(trackID string) (enums.PayoutStatus, error) {
	return enums.SUCCESS, nil
}
