package test

import (
	"context"
	"testing"

	"gil_teacher/app/consts"
	dao_task "gil_teacher/app/dao/task"

	"github.com/stretchr/testify/assert"
)

func TestTask(t *testing.T) {
	TaskStatDao := dao_task.NewTaskReportDAO(db, Clog)
	stats, err := TaskStatDao.FindAll(context.Background(), 1, &consts.DBPageInfo{Page: 1, Limit: 10})
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestTaskAnswerCount(t *testing.T) {
	TaskAnswerCountDao := dao_task.NewTaskStudentDetailsDao(db, Clog)
	answerStat, err := TaskAnswerCountDao.GetTaskAnswerCountStat(context.Background(), 1, 1, []string{"1", "2", "3"})
	assert.NoError(t, err)
	assert.NotNil(t, answerStat)
}
