package httpcodec

import (
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	netHttp "net/http"
	"strconv"
	"time"
)

type HttpStandardResponse struct {
	Status       int32       `json:"status"`        //业务状态 1成功 其他失败
	Code         int32       `json:"code"`          //errors.code定义
	Msg          string      `json:"message"`       //错误/成功信息
	TraceId      string      `json:"trace_id"`      //错误/成功信息
	ResponseTime int64       `json:"response_time"` //响应时间
	Data         interface{} `json:"data"`          //业务数据
}

const (
	STAT_SUCCESS     = 1
	STAT_SUCCESS_MSG = "success"
)

//ErrorEncoderHandler 错误编码 errors metadata中获取stat, 作为顶级字段响应返回
func ErrorEncoderHandler(w netHttp.ResponseWriter, r *netHttp.Request, err error) {
	se := errors.FromError(err)
	//bodyStat, _ := strconv.Atoi(se.Metadata["stat"])
	bodyCode, _ := strconv.Atoi(se.Metadata["code"])
	response := &HttpStandardResponse{
		Status:       400,
		Code:         int32(bodyCode),
		Msg:          se.Message,
		Data:         struct{}{},
		ResponseTime: time.Now().Unix(),
	}

	codeObj, _ := http.CodecForRequest(r, "Accept")
	body, err := codeObj.Marshal(response)
	if err != nil {
		w.WriteHeader(netHttp.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/"+codeObj.Name())
	w.WriteHeader(int(se.Code))
	_, _ = w.Write(body)
}

//ResponseEncoderHandler  正确响应解码
func ResponseEncoderHandler(w netHttp.ResponseWriter, r *netHttp.Request, v interface{}) error {
	traceId := ""
	if header, ok := transport.FromServerContext(r.Context()); ok {
		traceId = header.RequestHeader().Get("x-traceid")
	}

	response := &HttpStandardResponse{
		Status:       200,
		Code:         0,
		Msg:          STAT_SUCCESS_MSG,
		TraceId:      traceId,
		Data:         v,
		ResponseTime: time.Now().Unix(),
	}

	codeObj, _ := http.CodecForRequest(r, "Accept")
	data, err := codeObj.Marshal(response)
	//json.MarshalOptions.Marshal(response)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/"+codeObj.Name())
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}
