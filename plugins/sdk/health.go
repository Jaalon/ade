package sdk

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ade/plugins-sdk/contract"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type healthService struct {
	getStatus func() contract.HealthStatusEnum
}

func (h *healthService) serveHTTP(w http.ResponseWriter, req *http.Request) {
	status := contract.HealthStatus{
		Status:    h.getStatus(),
		Message:   "ok",
		Timestamp: time.Now().Unix(),
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "failed to encode health", http.StatusInternalServerError)
	}
}

func (h *healthService) checkGRPC() (*contract.HealthCheckResponse, error) {
	protoStatus, ok := mapHealthStatus(h.getStatus())
	if !ok {
		return nil, status.Error(codes.Internal, "unknown health status")
	}
	return &contract.HealthCheckResponse{
		Status:    protoStatus,
		Message:   "ok",
		Timestamp: time.Now().Unix(),
	}, nil
}

func mapHealthStatus(s contract.HealthStatusEnum) (contract.HealthCheckResponse_Status, bool) {
	switch s {
	case contract.HealthHealthy:
		return contract.HealthCheckResponse_HEALTHY, true
	case contract.HealthDegraded:
		return contract.HealthCheckResponse_DEGRADED, true
	case contract.HealthUnhealthy:
		return contract.HealthCheckResponse_UNHEALTHY, true
	default:
		return contract.HealthCheckResponse_UNKNOWN, false
	}
}
