//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"gil_teacher/app/conf"
	cLog "gil_teacher/app/core/logger"
	daoProvider "gil_teacher/app/dao/providers"
	"gil_teacher/app/domain"
	"gil_teacher/app/domain/behavior"
)

// 不需要 http rpc 等服务
// wireApp init kratos application.
func wireApp(
	serverConf *conf.Server,
	cnf *conf.Conf,
	data *conf.Data,
	config *conf.Config,
	logger log.Logger,
) (*behavior.BehaviorHandler, func(), error) {
	panic(wire.Build(
		cLog.ProviderSet,
		// middlewareProvider.ServerProviderSet,
		daoProvider.RepoProviderSet,
		domain.DomainProviderSet,
		// serviceProvider.ServiceProviderSet,
		// coreProvider.CoreProviderSet,
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
