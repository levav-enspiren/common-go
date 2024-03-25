package gormquery

import (
	"context"
	"encoding/json"
	"time"

	"gitea.greatics.net/common-go/gormquery/queryService"
	"gitea.greatics.net/common-go/skmap"
)

type QueryServiceModel struct {
	GrpcClient     queryService.QueryServiceClient
	RequestTimeout time.Duration
}

func (m *QueryServiceModel) Get(ctx context.Context, options skmap.Map) (results []skmap.Map, totalCount uint64, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(m.RequestTimeout))
	defer cancel()
	optionsBytes, err := json.Marshal(options)
	if err != nil {
		return
	}
	response, err := m.GrpcClient.Get(ctx, &queryService.OptionRequest{
		Options: optionsBytes,
	})
	if err != nil {
		return
	}
	err = json.Unmarshal(response.Results, &results)
	if err != nil {
		return
	}
	totalCount = response.TotalCount
	return
}

func (m *QueryServiceModel) Create(ctx context.Context, options skmap.Map) (result skmap.Map, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(m.RequestTimeout))
	defer cancel()
	optionsBytes, err := json.Marshal(options)
	if err != nil {
		return
	}
	response, err := m.GrpcClient.Create(ctx, &queryService.OptionRequest{
		Options: optionsBytes,
	})
	if err != nil {
		return
	}
	err = json.Unmarshal(response.Result, &result)
	return
}

func (m *QueryServiceModel) Update(ctx context.Context, options skmap.Map) (err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(m.RequestTimeout))
	defer cancel()
	optionsBytes, err := json.Marshal(options)
	if err != nil {
		return
	}
	_, err = m.GrpcClient.Update(ctx, &queryService.OptionRequest{
		Options: optionsBytes,
	})
	if err != nil {
		return
	}
	return
}

func (m *QueryServiceModel) Delete(ctx context.Context, options skmap.Map) (err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(m.RequestTimeout))
	defer cancel()
	optionsBytes, err := json.Marshal(options)
	if err != nil {
		return
	}
	_, err = m.GrpcClient.Delete(ctx, &queryService.OptionRequest{
		Options: optionsBytes,
	})
	if err != nil {
		return
	}
	return
}
