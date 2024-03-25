package grpchealth

import (
	"context"
	"errors"
)

type HealthService struct {
	HealthServer
}

func (a *HealthService) Check(ctx context.Context, request *HealthCheckRequest) (response *HealthCheckResponse, err error) {
	response = &HealthCheckResponse{
		Status: HealthCheckResponse_SERVING,
	}
	return
}

func (a *HealthService) Watch(_ *HealthCheckRequest, _ Health_WatchServer) (err error) {
	err = errors.New("not implemented")
	return
}
