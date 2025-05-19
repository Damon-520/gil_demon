//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"gil_teacher/app/conf"
	"gil_teacher/app/controller/providers"
	cLog "gil_teacher/app/core/logger"
	coreProvider "gil_teacher/app/core/providers"
	daoProvider "gil_teacher/app/dao/providers"
	"gil_teacher/app/domain"
	middlewareProvider "gil_teacher/app/middleware/provider"
	"gil_teacher/app/server"
	serviceProvider "gil_teacher/app/service/providers"
	thirdPartyProvider "gil_teacher/app/third_party/providers"
)

// wireApp init kratos application.
func wireApp(
	ctx context.Context,
	serverConf *conf.Server,
	cnf *conf.Conf,
	data *conf.Data,
	config *conf.Config,
	logger log.Logger,
) (*kratos.App, func(), error) {
	panic(wire.Build(
		cLog.ProviderSet,
		server.ServerProviderSet,
		middlewareProvider.ServerProviderSet,
		coreProvider.CoreProviderSet,
		daoProvider.RepoProviderSet,
		domain.DomainProviderSet,
		serviceProvider.ServiceProviderSet,
		providers.ControllerProviderSet,
		thirdPartyProvider.ThirdPartyProviderSet,
	))
}

// newElasticsearchClient creates a new Elasticsearch client.
func NewElasticsearchClient(config *conf.Elasticsearch) (*elasticsearch.Client, error) {
	esCfg := elasticsearch.Config{
		Addresses: []string{config.EsURL},
		Username:  config.Username,
		Password:  config.Password,
	}
	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, err
	}
	return client, nil
}
