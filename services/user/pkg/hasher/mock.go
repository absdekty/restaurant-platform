package hasher

type MockHasher struct {
}

func NewMock() *MockHasher {
	return &MockHasher{}
}

func (m *MockHasher) Hash(password string) (string, error) {
	return "hashed_" + password, nil
}

func (m *MockHasher) Compare(hashedPassword, password string) bool {
	return hashedPassword == "hashed_"+password
}
