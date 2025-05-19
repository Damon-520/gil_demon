package sidx

import "testing"

func Test_Sid(t *testing.T) {

	sid := NewSid()

	t.Log(sid.GenUint64())
	t.Log(sid.GenUint64())
	t.Log(sid.GenUint64())
}
