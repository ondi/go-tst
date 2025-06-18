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
	rnd := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), StringToUint64(t.Name())))
	for i := 1; i < 1_000_000; i++ {
		buf := GenerateString(rnd, 10+rnd.IntN(20), CHARSET)
		hx := StateSum64(0, buf)
		conflict, temp, size := storage.Add(hx, string(buf))
		if conflict {
			t.Fatalf("collision i=%v, hash=%v, storage=%q, buf=%q", i, hx, temp, buf)
		}
		if i%1_000_000 == 0 {
			t.Logf("i=%v, repeat=%v, hash=%v, storage=%v, buf=%q", i, repeat, hx, size, buf)
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
	{"XZtGF2$nW3YMqHpsTlo%PVj$", "Dmf85*ulAE#zMFdZCn9XQj/E4F"},
	{"u9saH51UD_j", "4T^piK~7S#*Ah$x$sHFiEPBd8"},
	{"dDxOSK&L#6IbW5Y#EN1WFInnE", "ds*DnW2&~D_m~n/"},
	{"WlocCY1@qfDCcJ33*gF", "E7S@S^_yn6nRwjj~/4p3"},
	{"1TL4@ZKAK1dTVwa", "4XGvSLF_Tju^MOa"},
	{"QI$%5axEMo2xC@", "l4Yw~tx@dJxFH77k@0lr2Om&"},
	{"8yj5LKpkJ^tZrDhW/c", "0/lBB~jie*NR@4"},
	{"zT1-^#Qxnt", "5g_eaPF#Skpq8"},
	{"xpwkc$q4rw0TY1g5#E#", "97ATATaHu*sMih@o7Ws8cOQm"},
	{"0I0iJEM8RQk5Xv%M3uKj@TpU3gUE8", "%m*xGOZ94Eu"},
}

func Test_Tst3_03(t *testing.T) {
	for _, v := range in {
		res1 := StateSum64(0, []byte(v[0]))
		res2 := StateSum64(0, []byte(v[1]))
		assert.Assert(t, res1 != res2, fmt.Sprintf("%v %q %q", res1, v[0], v[1]))
	}
}
