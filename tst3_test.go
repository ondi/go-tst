//
// go test -run Test_Tst3_02 -v -count=1 -timeout 0
//

package tst

import (
	"fmt"
	"hash/fnv"
	"io"
	"math/rand/v2"
	"sync"
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
func GenerateString(rnd *rand.Rand, length int, charset []byte) (out []byte) {
	out = make([]byte, length)
	for i := range length {
		out[i] = charset[rnd.IntN(len(charset))]
	}
	return
}

func StringToUint64(in string) uint64 {
	h := fnv.New64a()
	io.WriteString(h, in)
	return h.Sum64()
}

type Shard_t struct {
	Mx sync.Mutex
	Dm map[uint64]string
}

type Shards_t []Shard_t

func NewShards(shards int) (self Shards_t) {
	for i := 0; i < shards; i++ {
		self = append(self, Shard_t{Dm: map[uint64]string{}})
	}
	return
}

func (self Shards_t) Add(key uint64, value string) (conflict bool, value2 string, size int) {
	shard := key % uint64(len(self))
	// self[shard].Mx.Lock()
	// defer self[shard].Mx.Unlock()
	if value2, conflict = self[shard].Dm[key]; !conflict {
		self[shard].Dm[key] = value
	} else if value == value2 {
		conflict = false
	}
	size = len(self[shard].Dm)
	return
}

var storage = NewShards(1)

func test_02(t *testing.T) {
	t.Parallel()

	var repeat int
	salt := NewStateHash()
	rnd := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), StringToUint64(t.Name())))
	for i := 1; i < 1_000_000; i++ {
		salt.Reset()
		buf := GenerateString(rnd, 10+rnd.IntN(20), CHARSET)
		salted := salt.Sum64(buf)
		conflict, temp, size := storage.Add(salted, string(buf))
		if conflict {
			t.Fatalf("collision i=%v, salted=%v, storage=%q, buf=%q", i, salted, temp, buf)
		}
		if i%1_000_000 == 0 {
			t.Logf("i=%v, repeat=%v, salted=%v, storage=%v, buf=%q", i, repeat, salted, size, buf)
		}
	}
}

func Test_Tst3_02(t *testing.T) {
	for i := 0; i < 1; i++ {
		t.Run(fmt.Sprintf("test-%v", i), test_02)
	}
}

var in = [][]string{
	{"wc69qt5aD#m", "IgnB/zt0tIaY4y5WO"},
	{"cYp4W@ARg-v-ZE6zO", "Mn2s~HnQ~^2D9vAOCeUg&6EdG"},
	{"oJR%$L-ojMwDKLQRsgTZ%Ld4zJ", "RHqwScU9Bb/hAGS"},
	{"ieFOTMDRf26KDXp7X89HtKMZ", "TB@Fgg4kwvZ4qcj*e-k%tGW/o#v"},
	{"ZW@5A@PNeaVgWjxilsDg&", "~B-ApzTs$wL%"},
	{"yUR6QTdn*$-6-Mg", "kc^F^#trYSQpL2@YxrzsUG&kDa"},
	{"/*&IC&scY3lYVXh_s", "c~%qYSSCWVwyO*d2KMk2mzEL2x9*E"},
	{"bYK#ou-%/_lowdTJDtqk^FT&_dPa", "D/Z9OVZmQ#EQyBBC*Tl"},
	{"BKytNPwqzlEXB3C6ot*th#", "_qUwrmVY4f%wmqj"},
	{"24vWqxCezX76HWn^160", "/%UWIZSp@am^NG"},
	{"gIDFraDf&Xia-U@2&e", "5rkzBp53C_TkoD#r&jEXYhkuWR"},
}

func Test_Tst3_03(t *testing.T) {
	salt := NewStateHash()
	for _, v := range in {
		salt.Reset()
		res1 := salt.Sum64([]byte(v[0]))
		salt.Reset()
		res2 := salt.Sum64([]byte(v[1]))
		assert.Assert(t, res1 != res2, fmt.Sprintf("%v %q %q", res1, v[0], v[1]))
	}
}
