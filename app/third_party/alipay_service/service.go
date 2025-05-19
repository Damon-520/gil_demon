package alipay_service

import (
	"gil_teacher/app/conf"
	"gil_teacher/libs/httpx"

	"github.com/go-kratos/kratos/v2/log"
)

type AlipayService struct {
	httpClient *httpx.HttpClient
	conf       *conf.Config
	log        *log.Helper
}

func NewAlipayService(conf *conf.Config, logger log.Logger) IAlipayService {
	return &AlipayService{
		httpClient: httpx.NewHttpClient(logger),
		conf:       conf,
		log:        log.NewHelper(log.With(logger, "x_module", "util/NewAlipayService")),
	}
}
