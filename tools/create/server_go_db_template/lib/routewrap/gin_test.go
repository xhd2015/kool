package routewrap

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/xhd2015/xgo/support/assert"
)

// TestOptions defines options for testing a handler
type TestOptions struct {
	// Handler function to test
	Handler interface{}
	IDType  reflect.Type

	// HTTP method (GET, POST, etc.)
	Method string

	// Route path to register (e.g., "/user/:id")
	RoutePath string

	// Request path to execute (e.g., "/user/123?name=test")
	RequestPath string

	// Request body for POST/PUT requests
	Body string

	// Request headers
	Headers map[string]string

	// Session data for tests that need session
	SessionData map[string]interface{}

	// Expected response data
	ExpectedData interface{}

	// For error tests
	IsErrorTest        bool
	ExpectedStatusCode int
	ExpectedErrCode    float64
	ExpectedErrMsg     string
}

// MockSession implements the Session interface for testing
type MockSession struct {
	data map[string]interface{}
}

func NewMockSession(data map[string]interface{}) *MockSession {
	return &MockSession{data: data}
}

func (m *MockSession) Keys() []string {
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

func (m *MockSession) Get(key string) (interface{}, bool, error) {
	val, ok := m.data[key]
	return val, ok, nil
}

// MockSessionFactory implements the SessionFactory interface for testing
type MockSessionFactory struct {
	sessionData map[string]interface{}
	sessionKeys []*SessionKey
}

func NewMockSessionFactory(data map[string]interface{}) *MockSessionFactory {
	keys := make([]*SessionKey, 0, len(data))
	for k, v := range data {
		keys = append(keys, &SessionKey{
			Key:  k,
			Type: reflect.TypeOf(v),
		})
	}
	return &MockSessionFactory{
		sessionData: data,
		sessionKeys: keys,
	}
}

func (f *MockSessionFactory) SessionKeys() []*SessionKey {
	return f.sessionKeys
}

func (f *MockSessionFactory) GetSession(ctx *gin.Context) ISession {
	return NewMockSession(f.sessionData)
}

// runTest executes a test with the given options
func runTest(t *testing.T, opts TestOptions) {
	t.Helper()
	// Set defaults
	if opts.Method == "" {
		opts.Method = "GET"
	}

	// Setup session if needed
	var sessionFactory SessionFactory
	if opts.SessionData != nil {
		sessionFactory = NewMockSessionFactory(opts.SessionData)
	}

	// Setup router
	gin.SetMode(gin.TestMode)
	r := gin.New()
	processor := &Processor{
		SessionFactory: sessionFactory,
		IDType:         opts.IDType,
	}

	// Register route with handler
	switch opts.Method {
	case "GET":
		r.GET(opts.RoutePath, processor.Gin(opts.Handler))
	case "POST":
		r.POST(opts.RoutePath, processor.Gin(opts.Handler))
	case "PUT":
		r.PUT(opts.RoutePath, processor.Gin(opts.Handler))
	case "DELETE":
		r.DELETE(opts.RoutePath, processor.Gin(opts.Handler))
	default:
		r.Any(opts.RoutePath, processor.Gin(opts.Handler))
	}

	// Create request
	var req *http.Request
	if opts.Body != "" {
		req = httptest.NewRequest(opts.Method, opts.RequestPath, strings.NewReader(opts.Body))
	} else {
		req = httptest.NewRequest(opts.Method, opts.RequestPath, nil)
	}

	// Set headers
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	// Execute request
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Parse response
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check response
	if opts.IsErrorTest {
		// Error test case
		if opts.ExpectedStatusCode != 0 && w.Code != opts.ExpectedStatusCode {
			t.Errorf("expected status %d, got %d", opts.ExpectedStatusCode, w.Code)
		}

		if opts.ExpectedErrCode != 0 && resp["code"].(float64) != opts.ExpectedErrCode {
			t.Errorf("expected error code %v, got %v", opts.ExpectedErrCode, resp["code"])
		}

		if resp["msg"].(string) != opts.ExpectedErrMsg {
			t.Errorf("expected error message '%s', got '%v'", opts.ExpectedErrMsg, resp["msg"])
		}
	} else {
		// Success test case
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		expected := map[string]interface{}{
			"code": float64(0),
			"data": opts.ExpectedData,
		}

		if diff := assert.Diff(expected, resp); diff != "" {
			t.Error(diff)
		}
	}
}

func TestGin_NoParams(t *testing.T) {
	runTest(t, TestOptions{
		Handler: func() (interface{}, error) {
			return map[string]interface{}{"message": "success"}, nil
		},
		RoutePath:    "/test",
		RequestPath:  "/test",
		ExpectedData: map[string]interface{}{"message": "success"},
	})
}

func TestGin_IDParams(t *testing.T) {
	type ID int64
	runTest(t, TestOptions{
		IDType: reflect.TypeOf(ID(0)),
		Handler: func(id ID) (interface{}, error) {
			return map[string]interface{}{
				"id": id,
			}, nil
		},
		RoutePath:    "/user/:id",
		RequestPath:  "/user/123",
		ExpectedData: map[string]interface{}{"id": int64(123)},
	})
}

func TestGin_StructParams(t *testing.T) {
	type Req struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	runTest(t, TestOptions{
		Handler: func(req Req) (interface{}, error) {
			return map[string]interface{}{
				"id":   req.ID,
				"name": req.Name,
			}, nil
		},
		RoutePath:    "/user",
		RequestPath:  "/user",
		Method:       "POST",
		Body:         `{"id":123,"name":"test"}`,
		Headers:      map[string]string{"Content-Type": "application/json"},
		ExpectedData: map[string]interface{}{"id": int64(123), "name": "test"},
	})
}

func TestGin_SessionParam(t *testing.T) {
	type UserID int64
	type Role string
	runTest(t, TestOptions{
		Handler: func(userID UserID, role Role) (interface{}, error) {
			return map[string]interface{}{
				"userID": userID,
				"role":   role,
			}, nil
		},
		RoutePath:   "/profile",
		RequestPath: "/profile",
		SessionData: map[string]interface{}{
			"userID": UserID(456),
			"role":   Role("admin"),
		},
		ExpectedData: map[string]interface{}{
			"userID": 456,
			"role":   "admin",
		},
	})
}

func TestGin_ErrorReturn(t *testing.T) {
	runTest(t, TestOptions{
		Handler: func() (interface{}, error) {
			return nil, fmt.Errorf("something went wrong")
		},
		RoutePath:      "/error",
		RequestPath:    "/error",
		IsErrorTest:    true,
		ExpectedErrMsg: "something went wrong",
	})
}

func TestGin_InvalidFunction(t *testing.T) {
	var panicErr interface{}

	func() {
		defer func() {
			panicErr = recover()
		}()
		processor := &Processor{}

		// Not a function
		notFunc := "not a function"
		processor.Gin(notFunc)
	}()

	if panicErr == nil {
		t.Errorf("The code did not panic with invalid function")
	}
}
