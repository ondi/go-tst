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
	for i := 0; i < length; i++ {
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
	{A: "%G5#iW@FHT-HWfJX3", B: "tspTbAY3h$WwDQKfz"},
	{A: "Efn-Ga-%h^UT_ZyXv", B: "kG4Xc49rTUx&0Yham"},
	{A: "a$KYERvylDF400N~bdgfRA^o4A", B: "9Wi-rT3ddT5hz"},
	{A: "H@dq/znBhoG%ZqL", B: "#XQwn8DIxIVSq-jxtp0PLl"},
	{A: "N59IuxVpyp8", B: "V-4bEIL%SN#E"},
	{A: "j1U*qQ--R3", B: "DlfXFHcgOy6s4MVGw0vNcHyaCKPp"},
	{A: "eAJjraAxl6DZsRS_7yZMfdkqj0P$F", B: "ZLxZ0Ofk%iLfkqcCZI"},
	{A: "NYxRswXq%rg_D@4b6rMw#", B: "IKQdkpW%pKGPOeCY"},
	{A: "l$5x9$4R&f3pi2uUJ-/mLUV", B: "8mZloglGIj*2r#FqfQ4jcCh"},
	{A: "U$F-7CLuOaSY^v-_R", B: "oejLo*7EROOrGMm_O5"},
	{A: "A^Q408Jq0jFKARR~3nm1T&", B: "NdvHE$**BnalbxA"},
	{A: "gMGpgSKyFawBt@ZTH8A9-tMwy", B: "k&Mz#uTPqGih"},
	{A: "R/Akt4%7Np$wO5EXb4Ax&RZS96ufw", B: "6ZNNax1LTx5N6ZI"},
	{A: "1fr#X/~kACUICqlvSNMSr", B: "Wb&~4U3WKObWmirwvrAdz&nw"},
	{A: "z9xvjN1-L-fcKYU2H2P", B: "kTfDfs~lpgz"},
	{A: "OQzyimP^mKyKX-Z7Q", B: "/iOsJHUMJa"},
	{A: "TRE6dSp0cZXuJe&", B: "Cv5_un0ddi"},
	{A: "rVnC3cbOw8pS#9u/KsImGu98", B: "p7t9kxYyzZVxDyna_fsO~t%MtHgz-"},
	{A: "7H8xD&CnskMDuUsuxIYShi$p4l&", B: "XJD_TSlSYeawf*j9xIMpNnims~j6J"},
	{A: "tM-heCgLC%%bMF", B: "s#FxNKsgIiukz4"},
	{A: "vO*uJl51pJYRJ%#1we", B: "@HDNO$JwNBOq%hbI"},
	{A: "5zISkfL@kwS^RRMcxZC-PR3ow", B: "HtAVY8kHThP"},
	{A: "lmhBbCbaak9umO6EdQD%1D", B: "ypDIoOQZB24y%s*lk3"},
	{A: "IrUkjAsGWk2", B: "xAJAzFS@~kv6peHX4PzPB2#OYyT"},
	{A: "BTbUg1y/m%8JQT", B: "PAXHk%5B3V1KM1*KwcM0Dhf"},
	{A: "k8^m#kzC3s9sq^6hBwsm2%a", B: "LCHBI#2cNxiW3ki/Vb0"},
	{A: "H&k/mVS~/icstqm/T#98N^~&u4", B: "W-AOv76AF_"},
	{A: "7^bUs5jGkxez-5", B: "WGZX6%5@3Jj#oChan"},
	{A: "xK&^bZvqIJ_^CX/", B: "~Zy1a-f_@fIyd3/"},
	{A: "5-p6GdzjeTtfoRXfC6^T", B: "&oN$*e4^Mz0@@6*D"},
	{A: "gbPxOtHiOLJKaLH@f2sm_MCmO1z", B: "XPZ_~Jl#i4kVSMo"},
	{A: "EGvs%fjwWQUmBs4~-fbwK$EJUkDiB", B: "EfxmTR-zuge"},
	{A: "hA/#UAqdxX-", B: "F$FBEVyMeiFREPcCRmNJnsl$#1Y"},
	{A: "FWLtvnwGVNZ8Tg7XQ", B: "I0hyAksuUOYS@drQLVTuU1r&c"},
	{A: "rwfu/P~tIFSjg6Hhg2Bp/TfQT", B: "qg*s/U4P8&bULTkf&Q*kv4NP"},
	{A: "jri#vt&%d@", B: "FRfW*Dtwyy$Y^tY-d%-2f9J*h$NN"},
	{A: "g%hWlUYBXJ2Cy36UZ3^5", B: "xJ-jn-BmLYd"},
	{A: "&kGua#40I73", B: "XLp%I*G0*DL"},
	{A: "gRAJcMr2b0EOqgMOrr", B: "cZ_2K02ReNPutzl1ZfDxoxdFbaUm8"},
	{A: "xzvOBAoxl&upRc&O2gdr$07", B: "QsAJnqQJjQZax"},
	{A: "V74*XWIpboLsp_laNBRr0bbYD%", B: "X&PCLHI3@zVGQJNUae$X"},
	{A: "nfg*u~2&%*@$lUTWQQ/8", B: "8KQ6hHz0X//QmC0i0L*N"},
	{A: "RMV/2QMn%9FSN#gX8p3", B: "5qnRLiSvS#3m~&IZdEj"},
	{A: "Lr79pkaFv&pgDlFxbKce9aGXe", B: "53ScqZakIdzFu5&-IY"},
	{A: "GWPnCq/X&b@-WKXf^", B: "IgV~CeSNJg/1_M"},
	{A: "~9lUozGi#5~-jL-J", B: "^WCcFmxvbrPh"},
	{A: "MhM5cxXFMYGUQIH_zqx^nbo2", B: "jbLK/KOvmZb"},
	{A: "3f*k#~@Jf8h&s&9P*/BGA$01^M", B: "3*gRt_EV7w6z$RB@WdakoO"},
	{A: "EZSfINiHX4NY@C4ZXP-j~08sw_VS", B: "TUJd*PGPq_ZllI^~0vNil"},
	{A: "Kact#z5bH9nxkBe&", B: "p2ZJpiESj~s#Xnle$F6i5NAJ", Skip: true},
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

			var step1, step2 uint64
			var h1, h2 [1]uint64
			var a1, a2 [4]uint64

			for step1 < 256 {
				step1, h1[0], a1 = state1.Operation(step1, h1[0])
				step2, h2[0], a2 = state2.Operation(step2, h2[0])

				t.Logf("hash1=%v step1=%v a1=%v", h1[0], step1, a1)
				t.Logf("hash2=%v step2=%v a2=%v", h2[0], step2, a2)

				t.Logf("##### %v %v", h1[0] == h2[0], a1[0] == a2[0] || a1[1] == a2[1] || a1[2] == a2[2] || a1[3] == a2[3])
			}
			assert.Assert(t, v.Skip || v.Debug)
		}
	}
}

func Test_Tst3_05(t *testing.T) {
	in := "N*1Krz%2*sP&$#0nh"

	var state State256_t
	state.Reset()
	for _, code := range []byte(in) {
		state.StateNext(code)
	}
	res := state.Sum64()
	t.Logf("RES=%0X", res)
}
