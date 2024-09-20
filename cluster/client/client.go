package client

type TokenService interface {
	// AcquireFlowToken acquires tokens from remote token server, the checking logic is in local
	AcquireFlowToken(resource string, tokenCount uint32, statIntervalInMs uint32) (curCount int64, err error)
}

type nopTokenService struct {
}

func (n nopTokenService) AcquireFlowToken(_ string, _ uint32, _ uint32) (curCount int64, err error) {
	return 0, nil
}

func NopTokenService() TokenService {
	return nopTokenService{}
}