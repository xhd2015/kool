package routehelp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/xhd2015/xgo/support/assert"
)

func TestParseRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type testCase struct {
		name           string
		jsonBody       interface{}
		queryParams    map[string]string
		pathParams     []gin.Param
		requestStruct  interface{}
		expectedStruct interface{}
		expectSuccess  bool
	}

	type NestedStruct struct {
		Name  string `json:"name"`
		Value int64  `json:"value"`
	}

	tests := []testCase{
		{
			name:     "parse json body",
			jsonBody: map[string]interface{}{"name": "test", "value": int64(123)},
			requestStruct: &struct {
				Name  string `json:"name"`
				Value int64  `json:"value"`
			}{},
			expectedStruct: &struct {
				Name  string `json:"name"`
				Value int64  `json:"value"`
			}{
				Name:  "test",
				Value: 123,
			},
			expectSuccess: true,
		},
		{
			name: "parse query parameters",
			queryParams: map[string]string{
				"name": "test",
			},
			// Create an empty JSON body as the context always has a request body
			jsonBody: map[string]interface{}{},
			requestStruct: &struct {
				Name string `json:"name"`
			}{},
			expectedStruct: &struct {
				Name string `json:"name"`
			}{
				Name: "test",
			},
			expectSuccess: true,
		},
		{
			name: "parse path parameters",
			pathParams: []gin.Param{
				{Key: "id", Value: "123"},
			},
			// Create an empty JSON body as the context always has a request body
			jsonBody: map[string]interface{}{},
			requestStruct: &struct {
				ID int64 `json:"id"`
			}{},
			expectedStruct: &struct {
				ID int64 `json:"id"`
			}{
				ID: 123,
			},
			expectSuccess: true,
		},
		{
			name: "parse combined parameters",
			jsonBody: map[string]interface{}{
				"value": int64(456),
			},
			queryParams: map[string]string{
				"name": "test",
			},
			pathParams: []gin.Param{
				{Key: "id", Value: "123"},
			},
			requestStruct: &struct {
				ID    int64  `json:"id"`
				Name  string `json:"name"`
				Value int64  `json:"value"`
			}{},
			expectedStruct: &struct {
				ID    int64  `json:"id"`
				Name  string `json:"name"`
				Value int64  `json:"value"`
			}{
				ID:    123,
				Name:  "test",
				Value: 456,
			},
			expectSuccess: true,
		},
		{
			name: "path parameters override json body",
			jsonBody: map[string]interface{}{
				"id": int64(456),
			},
			pathParams: []gin.Param{
				{Key: "id", Value: "123"},
			},
			requestStruct: &struct {
				ID int64 `json:"id"`
			}{},
			expectedStruct: &struct {
				ID int64 `json:"id"`
			}{
				// The path parameter overrides the JSON body value
				ID: 123,
			},
			expectSuccess: true,
		},
		{
			name: "empty body with path params",
			pathParams: []gin.Param{
				{Key: "id", Value: "123"},
			},
			// Explicitly set an empty body
			jsonBody: map[string]interface{}{},
			requestStruct: &struct {
				ID int64 `json:"id"`
			}{},
			expectedStruct: &struct {
				ID int64 `json:"id"`
			}{
				ID: 123,
			},
			expectSuccess: true,
		},
		{
			name: "parse struct with private field",
			jsonBody: map[string]interface{}{
				"name": "test",
			},
			requestStruct: &struct {
				Name  string `json:"name"`
				value int64  // private field will be ignored
			}{},
			expectedStruct: &struct {
				Name  string `json:"name"`
				value int64  // private field will be ignored
			}{
				Name: "test",
			},
			expectSuccess: true,
		},
		{
			name: "parse struct with json tag omitted",
			jsonBody: map[string]interface{}{
				"name": "test",
			},
			pathParams: []gin.Param{
				{Key: "omitted", Value: "should not be set"},
			},
			requestStruct: &struct {
				Name    string `json:"name"`
				Omitted string `json:"-"` // omitted field
			}{},
			expectedStruct: &struct {
				Name    string `json:"name"`
				Omitted string `json:"-"` // omitted field
			}{
				Name: "test",
			},
			expectSuccess: true,
		},
		{
			name: "pointer to struct field",
			jsonBody: map[string]interface{}{
				"nested": map[string]interface{}{
					"name":  "test",
					"value": 123,
				},
			},
			requestStruct: &struct {
				Nested *NestedStruct `json:"nested"`
			}{},
			expectedStruct: &struct {
				Nested *NestedStruct `json:"nested"`
			}{
				Nested: &NestedStruct{
					Name:  "test",
					Value: 123,
				},
			},
			expectSuccess: true,
		},
		{
			name: "non-struct type",
			jsonBody: map[string]interface{}{
				"value": 123,
			},
			requestStruct: &map[string]interface{}{},
			expectedStruct: &map[string]interface{}{
				"value": json.Number("123"),
			},
			expectSuccess: true,
		},
		{
			name: "invalid type in path param",
			pathParams: []gin.Param{
				{Key: "id", Value: "not-a-number"},
			},
			// Need an empty JSON body to avoid nil request body
			jsonBody: map[string]interface{}{},
			requestStruct: &struct {
				ID int64 `json:"id"`
			}{},
			expectSuccess: false,
		},
		{
			name:          "non-pointer request struct",
			requestStruct: struct{}{},
			expectSuccess: false,
		},
		{
			name:          "nil request struct",
			requestStruct: nil,
			expectSuccess: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a response recorder and test context
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			// Create the HTTP request with JSON body if provided
			var body []byte
			var err error

			if tc.jsonBody != nil {
				body, err = json.Marshal(tc.jsonBody)
				if err != nil {
					t.Fatalf("failed to marshal JSON: %v", err)
				}
			} else {
				// Provide empty body to avoid nil request body
				body = []byte("{}")
			}

			httpReq := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
			httpReq.Header.Set("Content-Type", "application/json")

			// Add query parameters if any
			if len(tc.queryParams) > 0 {
				q := url.Values{}
				for k, v := range tc.queryParams {
					q.Add(k, v)
				}
				httpReq.URL.RawQuery = q.Encode()
			}

			ctx.Request = httpReq

			// Set path parameters if any
			if len(tc.pathParams) > 0 {
				ctx.Params = tc.pathParams
			}

			// Handle expected panics
			if tc.requestStruct == nil || tc.name == "non-pointer request struct" {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic for test case: %s", tc.name)
					}
				}()
			}

			// Call the function
			result := ParseRequest(ctx, tc.requestStruct)

			// Check the result
			if result != tc.expectSuccess {
				t.Errorf("ParseRequest() = %v, want %v", result, tc.expectSuccess)
			}

			// If success is expected, check the struct values
			if tc.expectSuccess {
				if diff := assert.Diff(tc.expectedStruct, tc.requestStruct); diff != "" {
					t.Errorf("Struct mismatch: %s", diff)
				}
			}
		})
	}
}
