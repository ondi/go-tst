//
//
//

package tst

import (
	"testing"

	"gotest.tools/assert"
)

func Test_Tst2_01(t *testing.T) {
	temp := &Tree2_t[string]{}

	temp.Add("/", "/")
	temp.Add("/debug", "/debug")
	temp.Add("/debug/size", "/debug/size")
	temp.Add("/debug/often", "/debug/often")
	temp.Add("/debug/metrics", "/debug/metrics")

	var found int
	var value string

	value, found = temp.Search("")
	assert.Assert(t, found == 0)

	value, found = temp.Search("v1/test")
	assert.Assert(t, found == 0)

	value, found = temp.Search("/debu")
	assert.Assert(t, value == "/")

	value, found = temp.Search("/v1/test")
	assert.Assert(t, value == "/")

	value, found = temp.Search("/debug/test")
	assert.Assert(t, value == "/debug")

	value, found = temp.Search("/debug/size")
	assert.Assert(t, value == "/debug/size")

	value, found = temp.Search("/debug/often")
	assert.Assert(t, value == "/debug/often")

	value, found = temp.Search("/debug/often/very")
	assert.Assert(t, value == "/debug/often")

	value, found = temp.Search("/debug/metrics")
	assert.Assert(t, value == "/debug/metrics")

	value, found = temp.Search("/debug/metrics2")
	assert.Assert(t, value == "/debug/metrics")
}
