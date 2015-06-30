package mockwriter

import ()

type MockWriter struct {
	Written     []byte
	returnError error
}

func New() *MockWriter {
	return &MockWriter{}
}

func (m *MockWriter) ReturnError(err error) {
	m.returnError = err
}

func (m *MockWriter) Write(p []byte) (n int, err error) {
	if m.returnError != nil {
		return 0, m.returnError
	}

	m.Written = p
	return len(m.Written), nil
}
