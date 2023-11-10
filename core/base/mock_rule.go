package base

type MockRule struct {
	Id   string `json:"id"`
	Name string `json:"name,omitempty"`
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

func (m *MockRule) RuleName() string {
	return m.Name
}
