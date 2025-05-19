package alipay_service

import "context"

type IAlipayService interface {
	OrderInfo(cxt context.Context) (OrderInfo, error)
}
