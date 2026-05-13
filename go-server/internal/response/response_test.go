package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAPIResponseStruct(t *testing.T) {
	resp := APIResponse{
		Success: true,
		Data:    map[string]string{"key": "value"},
		Message: "success",
		Code:    "OK",
	}

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if resp.Data == nil {
		t.Error("Expected Data to not be nil")
	}
	if resp.Message != "success" {
		t.Errorf("Expected Message to be 'success', got '%s'", resp.Message)
	}
	if resp.Code != "OK" {
		t.Errorf("Expected Code to be 'OK', got '%s'", resp.Code)
	}
}

func TestPaginatedResponseStruct(t *testing.T) {
	resp := PaginatedResponse{
		Success:  true,
		Data:     []string{"item1", "item2"},
		Total:    100,
		Page:     1,
		PageSize: 10,
	}

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if resp.Total != 100 {
		t.Errorf("Expected Total to be 100, got %d", resp.Total)
	}
	if resp.Page != 1 {
		t.Errorf("Expected Page to be 1, got %d", resp.Page)
	}
	if resp.PageSize != 10 {
		t.Errorf("Expected PageSize to be 10, got %d", resp.PageSize)
	}
}

func TestListResponseStruct(t *testing.T) {
	resp := ListResponse{
		Success: true,
		Data:    []int{1, 2, 3},
		Total:   3,
	}

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if resp.Total != 3 {
		t.Errorf("Expected Total to be 3, got %d", resp.Total)
	}
}

func TestErrorResponseStruct(t *testing.T) {
	resp := ErrorResponse{
		Success: false,
		Message: "error occurred",
		Code:    "ERR_001",
	}

	if resp.Success {
		t.Error("Expected Success to be false")
	}
	if resp.Message != "error occurred" {
		t.Errorf("Expected Message to be 'error occurred', got '%s'", resp.Message)
	}
	if resp.Code != "ERR_001" {
		t.Errorf("Expected Code to be 'ERR_001', got '%s'", resp.Code)
	}
}

func TestMFARequiredResponseStruct(t *testing.T) {
	resp := MFARequiredResponse{
		Success:     false,
		MFARequired: true,
		Message:     "MFA required",
		Code:        "MFA_REQUIRED",
	}

	if resp.Success {
		t.Error("Expected Success to be false for MFA required")
	}
	if !resp.MFARequired {
		t.Error("Expected MFARequired to be true")
	}
	if resp.Code != "MFA_REQUIRED" {
		t.Errorf("Expected Code to be 'MFA_REQUIRED', got '%s'", resp.Code)
	}
}

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	testData := map[string]int{"count": 42}
	Success(c, testData)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
}

func TestSuccessWithMessage(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	SuccessWithMessage(c, "Operation completed")

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if resp.Message != "Operation completed" {
		t.Errorf("Expected Message 'Operation completed', got '%s'", resp.Message)
	}
}

func TestSuccessPaginated(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := []string{"a", "b", "c"}
	SuccessPaginated(c, data, 100, 2, 10)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if resp.Total != 100 {
		t.Errorf("Expected Total 100, got %d", resp.Total)
	}
	if resp.Page != 2 {
		t.Errorf("Expected Page 2, got %d", resp.Page)
	}
	if resp.PageSize != 10 {
		t.Errorf("Expected PageSize 10, got %d", resp.PageSize)
	}
}

func TestSuccessList(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := []int{1, 2, 3}
	SuccessList(c, data, 3)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp ListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if resp.Total != 3 {
		t.Errorf("Expected Total 3, got %d", resp.Total)
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Error(c, http.StatusBadRequest, "invalid input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Success {
		t.Error("Expected Success to be false")
	}
	if resp.Message != "invalid input" {
		t.Errorf("Expected Message 'invalid input', got '%s'", resp.Message)
	}
}

func TestErrorWithCode(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ErrorWithCode(c, http.StatusNotFound, "resource not found", "NOT_FOUND")

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Success {
		t.Error("Expected Success to be false")
	}
	if resp.Message != "resource not found" {
		t.Errorf("Expected Message 'resource not found', got '%s'", resp.Message)
	}
	if resp.Code != "NOT_FOUND" {
		t.Errorf("Expected Code 'NOT_FOUND', got '%s'", resp.Code)
	}
}

func TestBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	BadRequest(c, "bad request message")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Unauthorized(c, "unauthorized message")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	NotFound(c, "not found message")

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Forbidden(c, "forbidden message")

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestInternalServerError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	InternalServerError(c, "internal error message")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestTooManyRequests(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	TooManyRequests(c, "rate limited")

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}
}

func TestServiceUnavailable(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ServiceUnavailable(c, "service down")

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

func TestMFARequired(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	MFARequired(c, "MFA required for this operation")

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for MFARequired, got %d", http.StatusOK, w.Code)
	}

	var resp MFARequiredResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Success {
		t.Error("Expected Success to be false for MFA required")
	}
	if !resp.MFARequired {
		t.Error("Expected MFARequired to be true")
	}
	if resp.Code != "MFA_REQUIRED" {
		t.Errorf("Expected Code 'MFA_REQUIRED', got '%s'", resp.Code)
	}
}

func TestSuccessWithNilData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Success(c, nil)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
}

func TestSuccessListWithEmptySlice(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	SuccessList(c, []string{}, 0)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp ListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if resp.Total != 0 {
		t.Errorf("Expected Total 0, got %d", resp.Total)
	}
}
