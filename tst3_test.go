//
// go test -run Test_Tst3_02 -v -count=1 -timeout 0
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

var CHARSET = []byte{
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '~', '@', '#', '$', '%', '^', '&', '*', '-', '_', '/',
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
}

// rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 1))
func GenerateString(rnd *rand.Rand, length int, charset []byte, out *bytes.Buffer) {
	// h := fnv.New64a()
	for range length {
		r := rnd.IntN(len(charset))
		b := charset[r : r+1]
		// h.Write(b)
		out.Write(b)
	}
	// return h.Sum64()
}

func Test_Tst3_02(t *testing.T) {
	// t.Parallel()

	var repeat int
	var buf bytes.Buffer
	storage := map[uint64]string{}
	salt := NewStateSalted()
	rnd := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 1))
	for i := 1; i < 1_000_000; i++ {
		buf.Reset()
		salt.Reset()
		GenerateString(rnd, 10+rnd.IntN(20), CHARSET, &buf)
		salted := salt.StateSalted(buf.Bytes())
		if i%1_000_000 == 0 {
			t.Logf("i=%v, repeat=%v, storage=%v, salted=%v, buf=%q", i, repeat, len(storage), salted, buf.Bytes())
		}
		if temp, ok := storage[salted]; ok {
			if temp == buf.String() {
				repeat++
			} else {
				t.Fatalf("collision salted=%v, storage=%v, buf=%q", salted, temp, buf.Bytes())
			}
		} else {
			storage[salted] = buf.String()
		}
	}
	t.Logf("storage=%v", len(storage))
}

var in = [][]string{
	// x = 85, y = 170
	// self.x = (self.x + 1) % 256
	// self.y = (self.y*int(self.state[self.x]) + int(in) + 1) % 256
	{"RpehpLGfEL", "VTehpLGf1x"},
	// x = 0, y = 0
	// self.x = (self.x + 1) % 256
	// self.y = (self.y*int(self.state[self.x]) + int(in) + 1) % 256
	{"pnc0zgYXPm8", "^WW$zgYXzM"},
	// x = 0, y = 1
	// self.x = (self.x + 1) % 256
	// self.y = (self.y*int(self.state[self.x]) + int(in) + 1) % 256
	{"%sc3e^tAkEE@", "ph~tp^tA8x"},
}

func Test_Tst3_03(t *testing.T) {
	salt := NewStateSalted()
	for i1, v1 := range in {
		for _, v2 := range v1 {
			salt.Reset()
			res := salt.StateSalted([]byte(v2))
			t.Logf("i=%v, res=%v, in=%v", i1, res, v2)
		}
	}
}
