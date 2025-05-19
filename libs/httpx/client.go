package httpx

import (
	"context"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-resty/resty/v2"
)

const (
	defaultTimeout = 3 * time.Second
)

type HttpClient struct {
	client  *resty.Client
	baseUrl string
	log     *log.Helper
}

func NewHttpClient(logger log.Logger, options ...DialOption) *HttpClient {

	do := dialOptions{}
	for _, option := range options {
		option.f(&do)
	}

	if do.timeout == 0 {
		do.timeout = defaultTimeout
	}

	client := resty.New().SetTimeout(do.timeout)
	return &HttpClient{
		client: client,
		log:    log.NewHelper(log.With(logger, "x_module", "data/NewHttpClient")),
	}
}

type dialOptions struct {
	timeout time.Duration
}

type DialOption struct {
	f func(options *dialOptions)
}

func (h *HttpClient) DialTimeout(d time.Duration) DialOption {
	return DialOption{func(do *dialOptions) {
		do.timeout = d
	}}
}

func DialTimeout(d time.Duration) DialOption {
	return DialOption{func(do *dialOptions) {
		do.timeout = d
	}}
}

func (h *HttpClient) SetBaseUrl(baseUrl string) *HttpClient {
	h.baseUrl = baseUrl
	return h
}

func (h *HttpClient) Get(ctx context.Context, path string, query, header map[string]string, response interface{}) error {
	resp, err := h.client.R().
		SetContext(ctx).
		SetQueryParams(query).
		SetHeaders(header).
		SetResult(&response).
		Get(path)

	if err != nil {
		h.log.WithContext(ctx).Errorf("httpx client err:%v", err)
		return err
	}

	if resp.StatusCode() != 200 {
		h.log.WithContext(ctx).Errorf("httpx get client status err:%v", resp.StatusCode())
		return errors.New("http请求失败")
	}
	return nil
}

func (h *HttpClient) Post(ctx context.Context, path string, body interface{}, header map[string]string, response interface{}) error {
	r := h.client.R().SetContext(ctx)
	if body != nil {
		r = r.SetBody(body)
	}
	if header != nil {
		r = r.SetHeaders(header)
	}

	resp, err := r.
		SetResult(response).
		Post(path)
	if err != nil {
		h.log.WithContext(ctx).Errorf("httpx client err:%v", err)
		return err
	}

	if resp.StatusCode() != 200 {
		h.log.WithContext(ctx).Errorf("httpx post client status err:%v", resp.StatusCode())
		return errors.New("http请求失败")
	}
	return nil
}
