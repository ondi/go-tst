//
// go test ./... -manual -v -count=1 -timeout 0
//

package tst

import (
	"flag"
	"fmt"
	"hash/fnv"
	"io"
	"math/rand/v2"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"gotest.tools/assert"
)

var flag_manual = flag.Bool("manual", false, "Запускать ручные тесты")
var collisions int

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

func test_02(t *testing.T, storage Shards_t, count int) {
	fd, err := os.OpenFile("collisions.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fd = os.Stderr
	} else {
		defer fd.Close()
	}

	var repeat int
	rnd := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), StringToUint64(t.Name())))
	for i := 1; i < count; i++ {
		value1 := GenerateString(rnd, 10+rnd.IntN(20), CHARSET)
		var state State256_t
		state.Reset()
		for _, code := range value1 {
			state.StateNext(code)
		}
		hx := state.Sum64()
		conflict, value2, size := storage.Add(hx, string(value1))
		if conflict {
			collisions++
			fmt.Fprintf(fd, "%s\t%v\t%q\t%q\n", t.Name(), i, value2, value1)
			t.Errorf("%v collision=%v i=%v, hash=%0X, value1=%q, valuw2=%q\n", t.Name(), collisions, i, hx, value1, value2)
		}
		if i%1_000_000 == 0 {
			t.Logf("%v i=%v, collision=%v, repeat=%v, storage=%v, hash=%0X, buf=%q", t.Name(), i, collisions, repeat, size, hx, value1)
		}
	}
}

func Test_Tst3_02(t *testing.T) {
	if flag_manual == nil || *flag_manual == false {
		t.Skip("skipped, add -manual to run")
	}
	for i := 0; i < 1_000; i++ {
		storage := NewShards(1)
		runtime.GC()
		t.Run(fmt.Sprintf("test-%v", i), func(t *testing.T) { test_02(t, storage, 700_000_000) })
	}
}

func Test_Tst3_03(t *testing.T) {
	storage := NewShards(1)
	t.Run("test-0", func(t *testing.T) { test_02(t, storage, 1_000_000) })
}

type DebugState_t struct {
	A     string
	B     string
	Debug bool
	Skip  bool
}

var in = []DebugState_t{
	{A: "6%qZjW33_o72CQwBsltF", B: "5KR5WmyG/J&z_M8hH"},
	{A: "lLD5B-DMn0oaljG78Aqg^b", B: "hIrdHtXM#87X_$iKMJ"},
	{A: "*gCBMAxWCIP%2YnKO*9I4W", B: "5YY$I@VK$5Djd-wEslLyQW/4"},
	{A: "Ig1HiBNVf2MO/F", B: "kL/@uK3sZ3_&GKebCT"},
	{A: "sw@&ZK^#by", B: "Jzf&ZK^#by"},
	{A: "jiFA5&nMvpOP1K", B: "1OT#YhDMvpOP1K"},
	{A: "iikgD2qmG/", B: "3B4eD2qmG/"},
	{A: "Ueh_M8qKvhnm0Y49/9B%Aj&VwI", B: "j#_B@MWopK"},
	{A: "RmPsw*HWdCOvK", B: "XwPnomWl^ZRScMo&6~GP9"},
	{A: "O0T_f92xgZPd/O_@", B: "Lu-&P71hPsHyqf$f"},
	{A: "%dBBzCq3WO", B: "eA@prdCZz20mUr@OvAE1"},
	{A: "cXUmas9lVc0Xz_IXabCMr3NymV", B: "FA%$o1D$ELYBaaGDBAcDVIC"},
	{A: "t&oXeZB*qz0#%%0t", B: "#_K#-hu$/s2V/-~t"},
	{A: "UEj3&G71/6CzCiorg", B: "@H/WHp9y2jiQensCwGWT_EX_il"},
	{A: "gG&2qwirJ6GGDLiH-bWY%7pRTAdpS", B: "4zDVz%c%$Z@8"},
	{A: "Dr5_r2^Ti0S1Z", B: "f2xWnyf4IFksld%j4svW5phJK"},
	{A: "6^/0uQ*rdiQg5k3y8lB&m", B: "I4g3LM4yfE@oEHeq/lE"},
	{A: "V0aUwgUlvo", B: "yQw1yJ3PMMb&o#Y"},
	{A: "pGyZaWDx#_S", B: "&JT&bt%IJsUjV@$i6@3M"},
	{A: "0DyZAIUSmvE-kWkrb13F$", B: "i^VBiJD_/HaXERkh*3BVD225^Gd2"},
	{A: "kq8OZPR3byKOjHRl42C*pJtq1u", B: "eIoNZUKWfy^z$D13V2M*5"},
	{A: "-_1oUep2V8sOhZ", B: "X/lP248A$Jl"},
	{A: "FZ~/7oQwqVbueR@3%RAm@UAC", B: "9Vylgk2f&ow@D@dnK79O3^D~G"},
	{A: "YK1We6poqz1YvY0gxAy&jbK", B: "juEm4@00d%L"},
	{A: "1mPXf9OM0x/*fGK65yUa/S&K", B: "ZX$J0oIl#AV8a%9pjbhEXD_*ds"},
	{A: "g%hiDoB@1k7aX4a", B: "ThvUO0kKI&9sUN%"},
	{A: "kG1MpgutW^^~EdiwS#9kCn", B: "E&QfVcv5*VL"},
	{A: "t53K0FNskElGMLLDICHmHmtr9o", B: "93-&k8Jl#PYhmCfX9_GfzC2Fx/mCM"},
	{A: "5HSJb//_YXZZ%UHdLEF^E", B: "68i3XotjZoRiw"},
	{A: "Rt-a$%s7cYgSfsH5", B: "0eY2Kwaw&g"},
	{A: "lbwWi@35Cpj5xDRPTKo&UDTPd%irr", B: "Qu-qIyt#Eb%V1AHkq4d5dCHHj4"},
	{A: "rM#4VwCB4&95RanX8t", B: "BFd4gHW^#Z~F1S_qL"},
	{A: "ianZs#NvKmyn/wEpO", B: "iKtN&m_&yk%PhZofG"},
	{A: "HzY8rz1k2H3RZvMNWJqu", B: "60dJ@wC&dIpT7fSr9dOCMiu9O0"},
	{A: "YO#_fSoW1%kHu", B: "NjPJ#NMr5-YqoFESmt&ubAl*u"},
	{A: "aptNp*_1JpgeS5VLi", B: "1fPN$3#waYuFU_hHg"},
	{A: "07ENtqKkkf0~spNbG", B: "G~kkZ7ladVOVSpAy%"},
	{A: "u7x3%fq~lVy^qcoOy", B: "rHTbR9J#@a6&SxhYG"},
	{A: "*v&G*eJJBkWC#bIn$", B: "$UkrsSP662NOO$Q0Q"},
	{A: "v*YK/ktPCg-OFOC3&EEdR3f", B: "Jq*mC8WT/3gJvs_zmwlFK"},
	{A: "4s8ICBi3kw92D9G82", B: "gPjDL%BI*AvXCdQ_w"},
	{A: "zw*&y-zYwPVPA~FeHLpDgcu8", B: "Vg/ZMlhBnf48ZXe^"},
	{A: "A7eds2PQOwQ0KoWHK", B: "JDPF8nWRbTgqdlKH2"},
	{A: "FiGxyz618gKk$#KxY&O#qK0*F", B: "3#vgwOgfV6e2V~ohr"},
	{A: "hyB#RBT5RSWReQF3#U", B: "$l5qt%MRAnk#e2g*XNP"},
	{A: "2IvOpyHNUXpN#", B: "7yQwNSlQl_w"},
	{A: "47Yzc~N1thZLr6cur", B: "C#Ylsf@1o1GtQ%ztg"},
	{A: "b4EUQ-YshHrj1fBF_4JhzW", B: "SJtSDNHPcSqmhvMPpfqR8#_qwe2bm"},
	{A: "2Rh5MKNCEmKOGsthH", B: "Pm0UdIpLz6ZQM"},
	{A: "RpJu6$5-H1/5EGV&c*ARM99", B: "uJoKrr8jZUi0Xp&iP~K1SPYxM^D"},
	{A: "-ZrhMT*Dm6CTN", B: "~7JjL*FlyE3M$zuhJeR8WWUc0nf2"},
	{A: "lBbH6-lbOchyjXa*zH8Php$kDt^~", B: "F6P5U#JQjT"},
	{A: "ioFY#Oh9DEY-M#2LTJOh~%^I_", B: "E8A/BAyE2J^c^4HhEe"},
	{A: "%G5#iW@FHT-HWfJX3", B: "tspTbAY3h$WwDQKfz", Debug: true},
	{A: "Efn-Ga-%h^UT_ZyXv", B: "kG4Xc49rTUx&0Yham"},
	{A: "a$KYERvylDF400N~bdgfRA^o4A", B: "9Wi-rT3ddT5hz"},
}

func Test_Tst3_04(t *testing.T) {
	for _, v := range in {
		var state1, state2 State256_t
		state1.Reset()
		state2.Reset()
		for _, code := range []byte(v.A) {
			state1.StateNext(code)
		}
		for _, code := range []byte(v.B) {
			state2.StateNext(code)
		}
		res1 := state1.Sum64()
		res2 := state2.Sum64()

		if res1 == res2 || v.Debug {
			same := map[int]int{}
			diff := map[int]int{}
			for i := 0; i < 256; i++ {
				if state1.state[i] == state2.state[i] {
					same[i/8]++
				} else {
					diff[i/8]++
				}
			}

			t.Logf("res1=%v\tlen1=%v\ta1=%v\tb1=%v\tin1=%q", res1, len(v.A), state1.a, state1.b, v.A)
			t.Logf("res2=%v\tlen2=%v\ta2=%v\tb2=%v\tin2=%q", res2, len(v.B), state2.a, state2.b, v.B)
			t.Logf("same = %v %v", len(same), same)
			t.Logf("diff = %v %v", len(diff), diff)

			var h1, h2 [4]uint64
			var a1, a2, o1, o2 [4]uint64

			for i := uint64(0); i < 256; i += 32 {
				a1[0] = state1.Uint64LE(state1.a + 1 + i)
				o1[0], h1[0] = state1.Operation(o1[3], h1[3], a1[0])
				a1[1] = state1.Uint64LE(state1.a + 9 + i)
				o1[1], h1[1] = state1.Operation(o1[0], h1[0], a1[1])
				a1[2] = state1.Uint64LE(state1.a + 17 + i)
				o1[2], h1[2] = state1.Operation(o1[1], h1[1], a1[2])
				a1[3] = state1.Uint64LE(state1.a + 25 + i)
				o1[3], h1[3] = state1.Operation(o1[2], h1[2], a1[3])

				a2[0] = state2.Uint64LE(state2.a + 1 + i)
				o2[0], h2[0] = state2.Operation(o2[3], h2[3], a2[0])
				a2[1] = state2.Uint64LE(state2.a + 9 + i)
				o2[1], h2[1] = state2.Operation(o2[0], h2[0], a2[1])
				a2[2] = state2.Uint64LE(state2.a + 17 + i)
				o2[2], h2[2] = state2.Operation(o2[1], h2[1], a2[2])
				a2[3] = state2.Uint64LE(state2.a + 25 + i)
				o2[3], h2[3] = state2.Operation(o2[2], h2[2], a2[3])

				t.Logf("a1=%v\ta2=%v\th1=%v\th2=%v\to1=%v\to2=%v\t%v", a1[0], a2[0], h1[0], h2[0], o1[0], o2[0], h1[0] == h2[0])
				t.Logf("a1=%v\ta2=%v\th1=%v\th2=%v\to1=%v\to2=%v\t%v", a1[1], a2[1], h1[1], h2[1], o1[1], o2[1], h1[1] == h2[1])
				t.Logf("a1=%v\ta2=%v\th1=%v\th2=%v\to1=%v\to2=%v\t%v", a1[2], a2[2], h1[2], h2[2], o1[2], o2[2], h1[2] == h2[2])
				t.Logf("a1=%v\ta2=%v\th1=%v\th2=%v\to1=%v\to2=%v\t%v", a1[3], a2[3], h1[3], h2[3], o1[3], o2[3], h1[3] == h2[3])

				t.Logf("#####")
			}
			assert.Assert(t, v.Skip || v.Debug)
		}
	}
}
