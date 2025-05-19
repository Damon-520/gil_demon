package test

import (
	"context"
	"gil_teacher/app/consts"
	"gil_teacher/app/dao/behavior"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBehavior(t *testing.T) {
	behaviorDAO := behavior.NewBehaviorDAO(chClients, Clog)
	session, err := behaviorDAO.GetCommunicationSession(context.Background(), "27d13511-4cd8-4ae3-8693-21f4bc458f29")
	assert.NoError(t, err)
	assert.NotNil(t, session)

	message, err := behaviorDAO.GetCommunicationSessionMessages(context.Background(), "27d13511-4cd8-4ae3-8693-21f4bc458f29", &consts.DBPageInfo{Page: 1, Limit: 10})
	assert.NoError(t, err)
	assert.NotNil(t, message)
}
