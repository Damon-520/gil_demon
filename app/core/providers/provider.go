package providers

import (
	"gil_teacher/app/core/kafka"
	"gil_teacher/app/core/zipkinx"

	"github.com/google/wire"
)

var CoreProviderSet = wire.NewSet(
	zipkinx.NewTracer,
	kafka.NewKafkaProducerClient,
)
