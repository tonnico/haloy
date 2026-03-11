package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/haloydev/haloy/internal/apitypes"
)

func TestHandleImagePrune_ReturnsDryRunPlan(t *testing.T) {
	s := newTestAPIServerForImages()
	s.imagePrune = func(context.Context, apitypes.ImagePruneRequest) (apitypes.ImagePruneResponse, error) {
		return apitypes.ImagePruneResponse{
			AppName:              "app",
			Keep:                 2,
			Applied:              false,
			RunningDeploymentIDs: []string{"20260311120000"},
			Tags: []apitypes.ImagePruneTag{
				{Tag: "app:20260311115900", DeploymentID: "20260311115900"},
			},
		}, nil
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/images/prune", strings.NewReader(`{"appName":"app","keep":2}`))
	rr := httptest.NewRecorder()

	s.handleImagePrune().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), `"applied":false`) {
		t.Fatalf("body = %q, want dry-run response", rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"runningDeploymentIds":["20260311120000"]`) {
		t.Fatalf("body = %q, want running deployment ids", rr.Body.String())
	}
}

func TestHandleImagePrune_ApplyFlagIsPassedThrough(t *testing.T) {
	s := newTestAPIServerForImages()
	var sawApply bool
	s.imagePrune = func(_ context.Context, req apitypes.ImagePruneRequest) (apitypes.ImagePruneResponse, error) {
		sawApply = req.Apply
		return apitypes.ImagePruneResponse{
			AppName: "app",
			Keep:    req.Keep,
			Applied: req.Apply,
		}, nil
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/images/prune", strings.NewReader(`{"appName":"app","keep":1,"apply":true}`))
	rr := httptest.NewRecorder()

	s.handleImagePrune().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !sawApply {
		t.Fatal("expected apply=true to be forwarded to image prune handler")
	}
	if !strings.Contains(rr.Body.String(), `"applied":true`) {
		t.Fatalf("body = %q, want apply response", rr.Body.String())
	}
}

func TestHandleImagePrune_RejectsNegativeKeep(t *testing.T) {
	s := newTestAPIServerForImages()
	req := httptest.NewRequest(http.MethodPost, "/v1/images/prune", strings.NewReader(`{"appName":"app","keep":-1}`))
	rr := httptest.NewRecorder()

	s.handleImagePrune().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}
