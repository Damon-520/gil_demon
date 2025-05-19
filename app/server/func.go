package server

import (
	"context"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

// 参考 gRPC gateway 的 DeafultHTTPErrorHandler
// panic if errHandle is nil
func runtimeHTTPErrorHandler(errHandle HTTPErrorHandleFunc) runtime.ErrorHandlerFunc {
	if errHandle == nil {
		panic("HTTPErrorHandleFunc is nil")
	}

	return func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
		// return Internal when Marshal failed

		s := status.Convert(err)
		pb := s.Proto()

		w.Header().Del("Trailer")
		w.Header().Del("Transfer-Encoding")

		contentType := marshaler.ContentType(pb)
		w.Header().Set("Content-Type", contentType)

		md, ok := runtime.ServerMetadataFromContext(ctx)
		if !ok {
			grpclog.Infof("Failed to extract ServerMetadata from context1")
		}

		handleForwardResponseServerMetadata(w, mux, md)

		// RFC 7230 https://tools.ietf.org/html/rfc7230#section-4.1.2
		// Unless the request includes a TE header field indicating "trailers"
		// is acceptable, as described in Section 4.3, a server SHOULD NOT
		// generate trailer fields that it believes are necessary for the user
		// agent to receive.
		var wantsTrailers bool

		if te := r.Header.Get("TE"); strings.Contains(strings.ToLower(te), "trailers") {
			wantsTrailers = true
			handleForwardResponseTrailerHeader(w, md)
			w.Header().Set("Transfer-Encoding", "chunked")
		}

		buf, st := errHandle(s)
		w.WriteHeader(st)
		if _, err := w.Write(buf); err != nil {
			grpclog.Infof("Failed to write response: %v", err)
		}

		if wantsTrailers {
			handleForwardResponseTrailer(w, md)
		}
	}
}

func defaultOutgoingHeaderMatcher(key string) (string, bool) {
	return fmt.Sprintf("%s%s", runtime.MetadataHeaderPrefix, key), true
}

func handleForwardResponseServerMetadata(w http.ResponseWriter, mux *runtime.ServeMux, md runtime.ServerMetadata) {
	for k, vs := range md.HeaderMD {
		// TODO 这里是写死的! 如果 ServerMux 使用了
		if h, ok := defaultOutgoingHeaderMatcher(k); ok {
			for _, v := range vs {
				w.Header().Add(h, v)
			}
		}
	}
}

func handleForwardResponseTrailerHeader(w http.ResponseWriter, md runtime.ServerMetadata) {
	for k := range md.TrailerMD {
		tKey := textproto.CanonicalMIMEHeaderKey(fmt.Sprintf("%s%s", runtime.MetadataTrailerPrefix, k))
		w.Header().Add("Trailer", tKey)
	}
}

func handleForwardResponseTrailer(w http.ResponseWriter, md runtime.ServerMetadata) {
	for k, vs := range md.TrailerMD {
		tKey := fmt.Sprintf("%s%s", runtime.MetadataTrailerPrefix, k)
		for _, v := range vs {
			w.Header().Add(tKey, v)
		}
	}
}
