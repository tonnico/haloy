package api

import (
	"context"
	"net/http"

	"github.com/haloydev/haloy/internal/apitypes"
	"github.com/haloydev/haloy/internal/docker"
	"github.com/haloydev/haloy/internal/logging"
)

func (s *APIServer) handleImagePrune() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req apitypes.ImagePruneRequest
		if err := decodeJSON(r.Body, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.AppName == "" {
			http.Error(w, "appName is required", http.StatusBadRequest)
			return
		}

		if req.Keep < 0 {
			http.Error(w, "keep must be at least 0", http.StatusBadRequest)
			return
		}

		if s.imagePrune != nil {
			resp, err := s.imagePrune(r.Context(), req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if err := encodeJSON(w, http.StatusOK, resp); err != nil {
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			}
			return
		}

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, defaultContextTimeout)
		defer cancel()

		cli, err := docker.NewClient(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cli.Close()

		plan, err := docker.PlanImagePrune(ctx, cli, req.AppName, "", req.Keep)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if req.Apply {
			logger := logging.NewLogger(s.logLevel, s.logBroker)
			if err := docker.ExecuteImagePrunePlan(ctx, cli, logger, plan); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		resp := apitypes.ImagePruneResponse{
			AppName:              req.AppName,
			Keep:                 req.Keep,
			Applied:              req.Apply,
			RunningDeploymentIDs: plan.RunningDeploymentIDs,
			Tags:                 make([]apitypes.ImagePruneTag, 0, len(plan.Tags)),
		}
		for _, tag := range plan.Tags {
			resp.Tags = append(resp.Tags, apitypes.ImagePruneTag{
				Tag:          tag.Tag,
				DeploymentID: tag.DeploymentID,
			})
		}

		if err := encodeJSON(w, http.StatusOK, resp); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
