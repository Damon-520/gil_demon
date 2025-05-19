package task

import (
	"context"
	"errors"

	"gil_teacher/app/controller/http_server/response"
	"gil_teacher/app/model/api"
)

// 获取学校指定班级指定科目的任务报告设置
func (h *TaskReportHandler) GetTaskReportSetting(ctx context.Context, schoolID, classID, subjectID int64) (*api.TaskReportSetting, error) {
	if schoolID == 0 || classID == 0 || subjectID == 0 {
		return nil, errors.New("schoolID, classID, subjectID is required")
	}

	setting, err := h.taskReportService.GetTaskReportSetting(ctx, schoolID, classID, subjectID)
	if err != nil {
		h.log.Error(ctx, "GetTaskReportSetting failed, schoolID:%d, classID:%d, subjectID:%d, error:%v", schoolID, classID, subjectID, err)
		return nil, err
	}
	return setting, nil
}

// 更新或设置任务报告设置
func (h *TaskReportHandler) UpdateTaskReportSetting(ctx context.Context, schoolID, teacherID int64, reportSetting *api.TaskReportSetting) *response.Response {
	if schoolID == 0 || teacherID == 0 || reportSetting == nil {
		return &response.ERR_INVALID_TASK_REPORT_SETTING
	}

	response := h.taskReportService.SaveTaskReportSetting(ctx, schoolID, teacherID, reportSetting)
	if response != nil {
		h.log.Error(ctx, "UpdateTaskReportSetting failed, schoolID:%d, teacherID:%d, reportSetting:%+v, error:%v",
			schoolID, teacherID, reportSetting, response)
		return response
	}
	return nil
}
