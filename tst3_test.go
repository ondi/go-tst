//
// go test -run Test_Tst3_02 -v -count=1
//

package tst

import (
	"bytes"
	"math/rand/v2"
	"testing"
	"time"

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

var CHARSET = []rune{
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '~', '@', '#', '$', '%', '^', '&', '*', '-', '_', '/',
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
}

// rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 1))
func GenerateString(rnd *rand.Rand, length int, charset []rune, out *bytes.Buffer) {
	for range length {
		out.WriteRune(charset[rnd.IntN(len(charset))])
	}
}

func Test_Tst3_02(t *testing.T) {
	var repeat int
	var buf bytes.Buffer
	storage := map[uint64]string{}
	salt := NewStateSalted()
	rnd := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 1))
	for i := 1; i < 1_000_000; i++ {
		buf.Reset()
		salt.Reset()
		GenerateString(rnd, 10+rnd.IntN(20), CHARSET, &buf)
		val := salt.StateSalted(buf.Bytes())
		if i%200_000 == 0 {
			t.Logf("i=%v, repeat=%v, storage=%v, sample %v %q", i, repeat, len(storage), val, buf.Bytes())
		}
		if temp, ok := storage[val]; ok {
			if temp == buf.String() {
				repeat++
				// t.Logf("repeat val=%v, temp=%q, buf=%q", val, temp, buf.Bytes())
			} else {
				t.Fatalf("collision val=%v, temp=%q, buf=%q", val, temp, buf.Bytes())
			}
		} else {
			storage[val] = buf.String()
		}
	}
	t.Logf("storage=%v", len(storage))
}
