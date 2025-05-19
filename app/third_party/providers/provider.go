package providers

import (
	"gil_teacher/app/third_party/alipay_service"
	"gil_teacher/app/third_party/middlewares/auth"
	"gil_teacher/app/third_party/sidx"
	"gil_teacher/app/third_party/volc_ai"
	"gil_teacher/app/third_party/zipkin_trace"

	"github.com/google/wire"
)

var ThirdPartyProviderSet = wire.NewSet(
	sidx.NewSid,
	auth.NewAdminAuth,
	alipay_service.NewAlipayService,
	zipkin_trace.NewZipkinTracer,
	volc_ai.NewClient,
)
