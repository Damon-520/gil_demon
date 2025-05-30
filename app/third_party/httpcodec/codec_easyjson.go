// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package httpcodec

import (
	json "encoding/json"
	kratosJson "github.com/go-kratos/kratos/v2/encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
	"google.golang.org/protobuf/proto"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson38c57360DecodeGit100talComKratosLibCodec(in *jlexer.Lexer, out *HttpStandardResponse) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "status":
			out.Status = int32(in.Int32())
		case "code":
			out.Code = int32(in.Int32())
		case "message":
			out.Msg = string(in.String())
		case "trace_id":
			out.Msg = string(in.String())
		case "response_time":
			out.ResponseTime = int64(in.Int64())
		case "data":
			if m, ok := out.Data.(easyjson.Unmarshaler); ok {
				m.UnmarshalEasyJSON(in)
			} else if m, ok := out.Data.(json.Unmarshaler); ok {
				_ = m.UnmarshalJSON(in.Raw())
			} else {
				out.Data = in.Interface()
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson38c57360EncodeGit100talComKratosLibCodec(out *jwriter.Writer, in HttpStandardResponse) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"status\":"
		out.RawString(prefix[1:])
		out.Int32(int32(in.Status))
	}
	{
		const prefix string = ",\"code\":"
		out.RawString(prefix)
		out.Int32(int32(in.Code))
	}
	{
		const prefix string = ",\"message\":"
		out.RawString(prefix)
		out.String(string(in.Msg))
	}
	{
		const prefix string = ",\"trace_id\":"
		out.RawString(prefix)
		out.String(string(in.TraceId))
	}
	{
		const prefix string = ",\"response_time\":"
		out.RawString(prefix)
		out.Int64(int64(in.ResponseTime))
	}
	{
		const prefix string = ",\"data\":"
		out.RawString(prefix)
		if m, ok := in.Data.(easyjson.Marshaler); ok {
			m.MarshalEasyJSON(out)
		} else if m, ok := in.Data.(json.Marshaler); ok {
			out.Raw(m.MarshalJSON())
		} else if m, ok := in.Data.(proto.Message); ok {
			out.Raw(kratosJson.MarshalOptions.Marshal(m))
		} else {
			out.Raw(json.Marshal(in.Data))
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v HttpStandardResponse) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson38c57360EncodeGit100talComKratosLibCodec(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v HttpStandardResponse) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson38c57360EncodeGit100talComKratosLibCodec(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *HttpStandardResponse) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson38c57360DecodeGit100talComKratosLibCodec(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *HttpStandardResponse) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson38c57360DecodeGit100talComKratosLibCodec(l, v)
}
