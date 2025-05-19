package domain

import (
	"gil_teacher/app/domain/behavior"
	"gil_teacher/app/domain/task"

	"github.com/google/wire"
)

var DomainProviderSet = wire.NewSet(
	behavior.NewBehaviorHandler,
	behavior.NewBehaviorProducer,
	behavior.NewSessionMessageHandler,
	task.NewTaskReportHandler,
)
