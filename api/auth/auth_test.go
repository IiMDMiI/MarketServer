package auth

import (
	"bytes"
	"errors"

	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IiMDMiI/MarketServer/pkg/dbservice"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDB struct {
	mock.Mock
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	ret := m.Called(query, args)
	return ret.Get(0).(sql.Result), ret.Error(1)
}
func (m *MockDB) QueryRow(query string, args ...interface{}) dbservice.SqlRow {
	ret := m.Called(query, args)
	return ret.Get(0).(dbservice.SqlRow)
}

func (m *MockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	ret := m.Called(query, args)
	return ret.Get(0).(*sql.Rows), ret.Error(1)
}
func (m *MockDB) Close() error {
	ret := m.Called()
	return ret.Error(0)
}

type MockSqlRow struct {
	mock.Mock
}

func (m *MockSqlRow) Err() error {
	ret := m.Called()
	return ret.Error(0)
}
func (m *MockSqlRow) Scan(dest ...any) error {
	ret := m.Called(dest)
	return ret.Error(0)
}

type MockResult struct {
	lastInsertId int64
	rowsAffected int64
}

func (m MockResult) LastInsertId() (int64, error) {
	return m.lastInsertId, nil
}

func (m MockResult) RowsAffected() (int64, error) {
	return m.rowsAffected, nil
}

func TestRegisterWhenUserNameDoesNotExist(t *testing.T) {
	dbservice.DB = &MockDB{}
	mockResult := MockResult{
		lastInsertId: 1,
		rowsAffected: 1,
	}
	dbservice.DB.(*MockDB).On("Exec", mock.Anything, mock.Anything).Return(mockResult, nil)

	//simulate user not found with
	//so that we can register new user
	mockSqlRow := &MockSqlRow{}
	mockSqlRow.On("Scan", mock.Anything).Return(errors.New("User not found"))

	dbservice.DB.(*MockDB).On("QueryRow", mock.Anything, mock.Anything).Return(mockSqlRow)

	user := User{
		Username: "testuser",
		Password: "testpassword",
	}

	payload, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	Register(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")
	expected := "User was registered"
	assert.Equal(t, expected, rr.Body.String(), "handler returned unexpected body")
}

func TestRegisterWhenUserNameAlreadyExists(t *testing.T) {
	dbservice.DB = &MockDB{}

	// simulate that the username already exists
	// so that we are unable to register a new user
	// nil means we found the username in db
	mockSqlRow := &MockSqlRow{}
	mockSqlRow.On("Scan", mock.Anything).Return(nil)

	dbservice.DB.(*MockDB).On("QueryRow", mock.Anything, mock.Anything).Return(mockSqlRow)

	user := User{
		Username: "testuser",
		Password: "testpassword",
	}

	payload, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	Register(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code, "handler returned wrong status code")
	expected := "Username already exists\n"
	assert.Equal(t, expected, rr.Body.String(), "handler returned unexpected body")
}

type MockPasswordProvider struct {
	mock.Mock
}

func (m *MockPasswordProvider) GetPassword(username string) ([]byte, error) {
	ret := m.Called(username)
	return ret.Get(0).([]byte), ret.Error(1)
}

func (m *MockPasswordProvider) ValidateHash(hashedPassword []byte, password []byte) error {
	ret := m.Called(hashedPassword, password)
	return ret.Error(0)
}

func TestLoginWithValidCredentials(t *testing.T) {
	dbservice.DB = &MockDB{}

	mockSqlRow := &MockSqlRow{}
	mockSqlRow.On("Scan", mock.Anything, mock.Anything).Return(nil)

	mockResult := MockResult{
		lastInsertId: 1,
		rowsAffected: 1,
	}
	dbservice.DB.(*MockDB).On("Exec", mock.Anything, mock.Anything).Return(mockResult, nil)
	dbservice.DB.(*MockDB).On("QueryRow", mock.Anything, mock.Anything).Return(mockSqlRow)

	user := User{
		Username: "testuser",
		Password: "testpassword",
	}

	payload, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	mockPassProvider := &MockPasswordProvider{}
	mockPassProvider.On("GetPassword", "testuser").Return([]byte("hashedpassword"), nil)
	mockPassProvider.On("ValidateHash", []byte("testpassword"), []byte("hashedpassword")).Return(nil)
	passProvider = mockPassProvider

	Login(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")
	assert.Contains(t, rr.Body.String(), "token", "handler returned unexpected body")
}

func TestLoginWithInvalidCredentials(t *testing.T) {
	dbservice.DB = &MockDB{}

	mockSqlRow := &MockSqlRow{}
	mockSqlRow.On("Scan", mock.Anything, mock.Anything).Return(errors.New("Invalid credentials"))

	dbservice.DB.(*MockDB).On("QueryRow", mock.Anything, mock.Anything).Return(mockSqlRow)

	user := User{
		Username: "testuser",
		Password: "wrongpassword",
	}

	payload, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	mockPassProvider := &MockPasswordProvider{}
	mockPassProvider.On("GetPassword", "testuser").Return([]byte("hashedpassword"), nil)
	mockPassProvider.On("ValidateHash", mock.Anything, mock.Anything).Return(errors.New("Invalid credentials"))
	passProvider = mockPassProvider

	Login(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "handler returned wrong status code")
	expected := "Invalid username or password\n"
	assert.Equal(t, expected, rr.Body.String(), "handler returned unexpected body")
}
