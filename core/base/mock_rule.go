package base

type MockRule struct {
	Id string `json:"id"`
}

func (m *MockRule) BlockType() BlockType {
	return BlockTypeFlow
}

func (m *MockRule) String() string {
	return "mock rule"
}

func (m *MockRule) ResourceName() string {
	return "mock resource"
}

func (m *MockRule) RuleID() string {
	return m.Id
}
