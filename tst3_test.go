//
// go test -run Test_Tst3_02 -v -count=1
//

package tst

import (
	"testing"

	"gotest.tools/assert"
)

func Test_Tst3_01(t *testing.T) {
	temp := NewTree3[string]()

	temp.Add("/", "/")
	temp.Add("/debug/kube", "/debug/kube")
	temp.Add("/debug", "/debug")
	temp.Add("/pprof", "/pprof")
	temp.Add("/metrics", "/metrics")

	var found int
	var value string

	value, _, found = temp.Search("")
	assert.Assert(t, found == 0)

	value, _, found = temp.Search("v1/test")
	assert.Assert(t, found == 0)

	value, _, found = temp.Search("/v1/test")
	assert.Assert(t, value == "/")

	value, _, found = temp.Search("/debu")
	assert.Assert(t, value == "/", value)

	value, _, found = temp.Search("/debug/test")
	assert.Assert(t, value == "/debug")

	value, _, found = temp.Search("/debug/size")
	assert.Assert(t, value == "/debug")

	value, _, found = temp.Search("/pprof/heap")
	assert.Assert(t, value == "/pprof")

	value, _, found = temp.Search("/pprof/profile")
	assert.Assert(t, value == "/pprof")

	value, _, found = temp.Search("/metrics/sql")
	assert.Assert(t, value == "/metrics")

	value, _, found = temp.Search("/metrics/page")
	assert.Assert(t, value == "/metrics")
}

func Test_Tst3_02(t *testing.T) {
	state := NewState8()
	for i := uint64(0); i < 10; i++ {
		res := ReplaceUint64(state, i)
		t.Logf("res(%v)=%v", i, res)
	}
	t.Logf("%+v", state)
}

func Test_Tst3_03(t *testing.T) {
	state := NewState8()
	in := []byte{0}
	for i := uint64(0); i < 10; i++ {
		in[0] = byte(i)
		in = state.Replace(in)
		t.Logf("res(%v)=%v", i, in)
	}
	t.Logf("%+v", state)
}
