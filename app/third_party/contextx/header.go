package contextx

import (
	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/transport"
	"strconv"
)

const (
	HeaderTokenKey         = "Token"           // token 名称
	HeaderClientTypeKey    = "Client"          // 客户端类型
	HeaderClientVerIdKey   = "Client-Ver-Id"   // 版本ID
	HeaderClientVerNameKey = "Client-Ver-Name" // 版本名称
	HeaderChannelKey       = "Channel"         // 注册渠道名称
	HeaderSourceKey        = "Source"          // 注册来源名称
	HeaderTimeKey          = "Time"            // 名称
	HeaderTraceId          = "x-traceid"       // traceId

)

type Header struct {
	Token         string
	ClientType    string // 终端
	ClientVerId   int32  // 9
	ClientVerName string // v1.0.0
	Channel       int32  // 渠道信息
	Source        int32  // 注册来源
	VersionInfo   string
	Time          int32 // 123455
	TraceId       string
}

type _header struct {
	ctx       *gin.Context
	ctxHeader transport.Header
	Header    *Header
}

func NewHeader(ctx *gin.Context) *_header {

	// var header _header
	// header.ctx = ctx
	// header.Header = &Header{}

	header := _header{
		ctx:    ctx,
		Header: &Header{}, // 申请内存
	}
	// 解析transport
	// if tr, ok := transport.FromServerContext(ctx); ok {
	//	header.ctxHeader = tr.RequestHeader()
	// }

	return &header
}

func (h *_header) GetAll() *Header {

	_ = h.GetToken()
	_ = h.GetClientType()
	_ = h.GetClientVerId()
	_ = h.GetClientVerName()
	_ = h.GetChannel()
	_ = h.GetSource()
	_ = h.GetTime()
	_ = h.GetTraceId()

	return h.Header
}

// GetToken 获取token
func (h *_header) GetToken() (token string) {

	h.Header.Token = h.ctx.GetHeader(HeaderTokenKey)
	return h.Header.Token
}

// GetClientType
// 终端 android(安卓端) ios(IOS) pad(Pad端) ipad(ipad端 program(小程序端) web(H5网页端)
func (h *_header) GetClientType() (client string) {

	h.Header.ClientType = h.ctx.GetHeader(HeaderClientTypeKey)

	return h.Header.ClientType
}

// GetClientVerId 版本自增ID 10
func (h *_header) GetClientVerId() (verId int32) {

	num, err := strconv.ParseInt(h.ctx.GetHeader(HeaderClientVerIdKey), 10, 32)
	if err != nil {
		return
	}

	h.Header.ClientVerId = int32(num)

	return h.Header.ClientVerId
}

// GetClientVerName 客户端版本名称 v1.8.0
func (h *_header) GetClientVerName() (verName string) {

	h.Header.ClientVerName = h.ctx.GetHeader(HeaderClientVerNameKey)

	return h.Header.ClientVerName
}

// GetChannel 获取渠道
func (h *_header) GetChannel() (channel int32) {

	num, err := strconv.ParseInt(h.ctx.GetHeader(HeaderChannelKey), 10, 32)
	if err != nil {
		return
	}
	h.Header.Channel = int32(num)

	return h.Header.Channel
}

// GetSource 获取注册来源
func (h *_header) GetSource() (sourceId int32) {

	num, err := strconv.ParseInt(h.ctx.GetHeader(HeaderSourceKey), 10, 32)
	if err != nil {
		return
	}
	h.Header.Source = int32(num)

	return h.Header.Source
}

// GetTime 获取时间戳
func (h *_header) GetTime() (time int32) {

	num, err := strconv.ParseInt(h.ctx.GetHeader(HeaderTimeKey), 10, 32)
	if err != nil {
		return
	}
	h.Header.Time = int32(num)

	return h.Header.Time
}

// GetTraceId 获取traceId
func (h *_header) GetTraceId() (traceId string) {

	h.Header.TraceId = h.ctx.GetHeader(HeaderTraceId)

	return h.Header.TraceId
}
