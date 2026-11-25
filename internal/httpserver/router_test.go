package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nabilfikrisp/cicd-demo/internal/modules/service/domain"
	"github.com/nabilfikrisp/cicd-demo/internal/modules/service/handler"
	"github.com/nabilfikrisp/cicd-demo/internal/modules/service/repository/memory"
	"github.com/nabilfikrisp/cicd-demo/internal/modules/service/usecase"
)

func TestAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := NewRouter(handler.NewServiceHandler(usecase.NewServiceUsecase(memory.NewServiceRepository())), "v1", "abc")
	request := func(method, path, body string, want int) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		if w.Code != want {
			t.Fatalf("%s %s: got %d, want %d: %s", method, path, w.Code, want, w.Body.String())
		}
		return w
	}
	request(http.MethodGet, "/health", "", http.StatusOK)
	version := request(http.MethodGet, "/version", "", http.StatusOK)
	if !strings.Contains(version.Body.String(), `"commit":"abc"`) {
		t.Fatalf("version: %s", version.Body.String())
	}
	request(http.MethodPost, "/services", `{"name":"billing","description":"invoices"}`, http.StatusCreated)
	request(http.MethodPost, "/services", `{"name":"billing"}`, http.StatusConflict)
	request(http.MethodPost, "/services", `{"name":" "}`, http.StatusBadRequest)
	request(http.MethodPost, "/services", `{`, http.StatusBadRequest)
	request(http.MethodPost, "/services", `{"name":"`+strings.Repeat("a", 16*1024)+`"}`, http.StatusRequestEntityTooLarge)
	w := request(http.MethodGet, "/services", "", http.StatusOK)
	var services []domain.Service
	if err := json.Unmarshal(w.Body.Bytes(), &services); err != nil || len(services) != 1 || services[0].Name != "billing" {
		t.Fatalf("list: %s, %v", w.Body.String(), err)
	}
}
