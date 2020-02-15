package cluster

type TokenService interface {
	RequestToken(ruleId int64, acquireCount uint32, prioritized bool) TokenResult
}


type TokenResult struct{
	Status int `json:"status"`
	Remaining int `json:"remaining"`
	WaitInMs int `json:"waitInMs"`

}
