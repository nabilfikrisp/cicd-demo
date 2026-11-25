package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nabilfikrisp/cicd-demo/internal/modules/service/domain"
)

type failingServices struct{ err error }

func (s failingServices) Create(context.Context, string, string) (domain.Service, error) {
	return domain.Service{}, s.err
}
func (s failingServices) List(context.Context) ([]domain.Service, error) { return nil, s.err }

func TestHandlerMapsUsecaseErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name string
		err  error
		want int
	}{
		{"invalid", domain.ErrInvalidService, http.StatusBadRequest},
		{"duplicate", domain.ErrServiceAlreadyExists, http.StatusConflict},
		{"unexpected", errors.New("sensitive storage details"), http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.POST("/services", NewServiceHandler(failingServices{err: tc.err}).Create)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/services", strings.NewReader(`{"name":"test"}`)))
			if w.Code != tc.want {
				t.Fatalf("status = %d, want %d", w.Code, tc.want)
			}
			if tc.want == http.StatusInternalServerError && strings.Contains(w.Body.String(), "sensitive") {
				t.Fatalf("internal error leaked: %s", w.Body.String())
			}
		})
	}
}
