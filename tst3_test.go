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
		var hx uint64
		value1 := GenerateString(rnd, 10+rnd.IntN(20), CHARSET)
		var state State256_t
		state.Reset()
		for _, code := range value1 {
			hx = Mix(hx, state.State(code, hx))
		}
		conflict, value2, size := storage.Add(hx, string(value1))
		if conflict {
			collisions++
			fmt.Fprintf(fd, "%s\t%v\t%q\t%q\n", t.Name(), i, value2, value1)
			t.Errorf("%v collision=%v i=%v, hash=%016X, value1=%q, value2=%q\n", t.Name(), collisions, i, hx, value1, value2)
		}
		if i%1_000_000 == 0 {
			t.Logf("%v i=%v, collision=%v, repeat=%v, storage=%v, hash=%016X, buf=%q", t.Name(), i, collisions, repeat, size, hx, value1)
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
	{A: "l@FVfTazm51", B: "QIR_#kv33D"},
	{A: "XTtKRclozF$VT_6TUXbZRIawl", B: "kBOfhk9cKa"},
	{A: "AUyfVlgIAMMrI9siGKxgvfZQZZCX", B: "1TKIysX_WT1o2P#w_HbvKlhQct"},
	{A: "EYF@iPT2E4l/%gm#l1c4gLUmvYa", B: "eGFGBSMX7ZsQ@@MLj5L"},
	{A: "Kact#z5bH9nxkBe&", B: "p2ZJpiESj~s#Xnle$F6i5NAJ"},
	{A: "@qQaFN_QkXIKD%mzJs1Q7Bq6my", B: "Ub2SSVgF5p2#8V-"},
	{A: "DH-k5LBP9#ERLqnar-Qi-1U", B: "H1kVt99rGG~OMsnJM#C&SSs"},
	{A: "mGEVUT0LeLt$IH8@vqght", B: "d*ST#LXkICq"},
	{A: "_-cBF6QT^8U9h7WZ", B: "sScSnQM&MzX~MEqjpC2JChc%IAkBA"},
	{A: "&APsgcW%Yx9tL^sCTo", B: "iij&vG-bvWVrQ9Jru0r/~VF"},
	{A: "RAIQGFj3YjDvk$P*NTzcB", B: "EGW7vqTDmQC~ZuamKHtY3V0"},
	{A: "uV_3cV2wzK_@ze0K", B: "sggD3oyyD*JzBPek^#qKiI3S"},
	{A: "zNqIOvJgoYz@UiWOjFYI~bqaLJMz-", B: "fNNxPN#D3j9T2$XE6"},
	{A: "C-jSLRKcCJB8", B: "Qx3$hUExNhL6"},
	{A: "Z*KstVax9e7FB", B: "%hv4TquwA0"},
	{A: "F*JyLRGRXscpA~piYL3", B: "J6$nDnIHz3q0Rmo"},
	{A: "7K8o$9IX%bXsOz#", B: "ZW/CrC~^ezqC"},
	{A: "U$wZ&O$%3Y_ZsSR$", B: "6dEG8h*%Ou6rKSP/G$pVJ6qbVFq^"},
	{A: "_K3#fFKl4zzt_XO829", B: "1G*e/1ieisiKKbXgMQOU%7JFe0aSX"},
	{A: "kP/7Hl49S1fW/O", B: "Z-uHJKouS1fwt/~Tut0~RhmpzBo9K"},
	{A: "Pz0stSU9wAl6l$83OgDSu7THpub", B: "glzWfFe7yHR2jOfQ"},
	{A: "lga#b/b59OS8", B: "UoeyDAIGI4"},
	{A: "owjKti0qQHZKUY$4P686Ea#nTp-", B: "zo5i^*SE0s$#UHzWd@rU"},
	{A: "fQ9fxSQ4f*Dl5Z&u", B: "GtF9DsTr21#"},
	{A: "VETvRNEd^rs0Y8Cin23mbpBpr", B: "Rg/thKA-vKICngaT6ZbA&k4^2b"},
	{A: "A83^0q&R4_dti", B: "q8JYno6RvMKhmFrCWWZSXql"},
	{A: "EOrbuj@m63SbSw2Li3Yp_E", B: "1uzln^H5PO&2*IBS2Sxnh_O"},
	{A: "R6~UZJ5bJ&r", B: "iYB8w$#3yoo%6xk5kLhZDiFv"},
	{A: "-cE9*tyFh9TW6dl8$RzPD", B: "OY_2bwr4C$K"},
	{A: "cKj29132e6l%t8Y2FW", B: "hyD/QXrH*DC4v5H"},
	{A: "5X4ekqNNL$fPLf46lPh", B: "o8dJJ@u~wTXbjKWW"},
	{A: "9fz59c5fm/2", B: "rCAlY3^SSnA"},
	{A: "d86XezT^O#e6VI#/a&tV", B: "c$r$7pKFSbFYMf2w~UMJ0g1vO/POX"},
	{A: "p_L/xZ&kDDE*Ny~9", B: "q_7Iv/2cX^srG1nHQHJ"},
	{A: "AlN#JbCv%442800Yc", B: "wBNACZBGOkdW1hL^imljtNtoxgq0"},
	{A: "EDyif_%kAclGu_SVBPBZQEqtza/L4", B: "bcuM3LCpb%QJVPr#&TBR3P2K%"},
	{A: "7KvgpO6uhf41R^@KxlO06Q", B: "jg$VTqm0CsvH8"},
	{A: "$msTojH*V/uNSRNd1_Ki2", B: "7WQ@1nuejXePtR"},
	{A: "sSA22#bO1oV8XOUBt1g", B: "m$lDFYQi%JckQ"},
	{A: "11i&KauCzl*g-G^", B: "cEM3RagxkyruN4S@9"},
	{A: "3hqBcY#QNJU$jg0*jdSrc*bS8uO", B: "nKFfA8NRz8k3V/uaE"},
	{A: "5qQeAnAx#p$@Ei4i8Q8", B: "@_nHQ$d02x_4MrXFz-p9qt@z3mzq"},
	{A: "V%_wwDdPJfL^EB@QS", B: "w$g%sCGPRq33L$z5$XC_9m6k"},
	{A: "SOzZdVre9huVeLDXeUx8Q", B: "uAJ%sSwyP-nOr"},
	{A: "-OP6zaqoxLPg_lRh$7ZWsNVjdD", B: "_tZTWWpM*7"},
	{A: "37X2~bQnEXgZ62P8VNi5", B: "NPh2glhHyUNdc%L$Rxy%_uL"},
	{A: "#J#xmbnAsY-Dt7", B: "/swPiivuTo9"},
	{A: "uNrI8ss_/@-kiC@ld8HW3@w@8Kw", B: "jfk1&aXSsT#"},
	{A: "kJ&-AtMSDf3XoWCf%70CBFq@^c", B: "aKp9GN^psfKK0f@al"},
	{A: "hAJqMC$eed$#*wj8BnarXw8mo", B: "txzGAaHdvBE"},
	{A: "koUH3oCbnwWa", B: "5^SPMxvf49RH^5HT20qVR/Z"},
	{A: "$S$qCIGAYac-E8e0#", B: "f~Z7MzfpbA9ympQEm"},
	{A: "ypYpcZL64Gl*5Dly$3qMQPXl", B: "IUKwr#iaJbWroqGguj0iB"},
	{A: "dY42uD68SbBnHgNKVZ-S$", B: "R6K2cLYLLoNR44"},
	{A: "yMYfNVwGdDCxx4", B: "n4$n52DJ4mvICLob3~x~3kW-"},
	{A: "ln$cz2Y5GD_R_k1D", B: "~9_zw$1_Mq#%asny&%eofxcFR4I"},
	{A: "RWrgTuRs15W99F", B: "UUncoJCSXE"},
	{A: "5-NO6$E0jx^S&Vft-bbsL_B", B: "%*mXq84dMncceit9%"},
	{A: "NT3R-XfjxTricu5XEro8", B: "%oLJatSL7j&X0q0E^I~g"},
	{A: "pAz1tb_aTBhD8qy0", B: "zjxR4FVoK4VUXs1wpwMadp@RxuAKE"},
	{A: "kwlIjN%Eui7*IeCK88hEbQrgw^", B: "&8HSdITojCbJ2BirCwlsO"},
	{A: "nihzm-_48jtVpcLx~pZqlgae", B: "J7gwfB/56w/$*l6Pp"},
	{A: "K6oa5PXg6iRE601XQYSVx^", B: "l6Xei2tM1Xl^rn/1d4~"},
	{A: "*cR@_h~-AEu*942wb-*W", B: "0E0bKx^vCstnCMG~~f-O-_nYBd*m9"},
	{A: "ecA/xJ^MJDoI*#UmGpuU5OiqDtbSQ", B: "qHMDmfS/jzeENjg"},
	{A: "B-qhSXX#Vj4Z8EtGz-~1", B: "*bPZ*5A8QbVTYq%Su4*"},
	{A: "mieKAOs41r@cyF4o%23W&#cUx6Pf", B: "RUX*CoiKG^&&K0*y4Xx"},
	{A: "#TsS4Q@k72l%Cm9l&s&pIon@N_7", B: "x_v7gzHAj/_ZHmif5cGJW8-xa"},
	{A: "IrWo_U-%VNP/h$u%cG3", B: "M~g^LP2crwTIRXGB*O&G58Y#^eJ^"},
	{A: "$a6Pw_r&%*/22", B: "oQYi/7wz33@ePL3m*/G2gT"},
	{A: "m4XG0f_Bb~tcamZKP_6", B: "vHGTRfsQHg1r1ZY0"},
	{A: "-/cuMb6^-&s$XY", B: "DK852sjm9IC6lk"},
	{A: "UH1Kzl&Y2A8R", B: "$olc_kZ3zK^SA-pZ"},
	{A: "quvc@*U@7qgJ8njP-ZXgV14", B: "rYboOI/Q#y5"},
	{A: "23TvcZ9GNz", B: "CX@m-ZGK*_"},
	{A: "@MaXLLft1zuCi%48l24aKW", B: "69~Qmuq2WEg_x6Wa3LVDpm"},
	{A: "0FDrCteUT9Jr", B: "#$@MlnX*UuHjUa"},
	{A: "@OiFIX-sIpN4", B: "bXqWhi~dyMa&EE&"},
	{A: "UoU2ea#Mt_8bLA^gGsE5#k52nzKd", B: "nFoRO#joM@8VJpRZa0wCB#@"},
	{A: "SpvN-_A&5PWk6pQ_a-vEW6", B: "z4nWh8Dp5JksOKB6J"},
	{A: "Cii2S$Ob4&S8LKnsDyl-UHufGa", B: "F%758_PcWR8^WrxmDMg"},
	{A: "Vld/iUuKGGF$YF#79dZ0m&Te%g", B: "osXBsudPI*XQ*GpMUv^N$Mky"},
	{A: "S~fDD*j&Nb&$AidgNcG", B: "IFt@K1**hEn%laHk7qj%lJc1ym#30"},
	{A: "_tlUsVR4YV79X9ZWI37yn", B: "VF3&wO0uyp0Gb5u2QcZ8-jX4"},
	{A: "vtz6elfoLVhbyCH~%-K3F%O", B: "8eORK2GFCiztXSbpN*htoJ9LbXI-"},
	{A: "Ko@~$obTR~P_sKssN#K*", B: "t9NI#U232VGbM*h$O_Rj"},
	{A: "92Y^m@589$D&BRv7", B: "Ft/KHHze6shnnQ~3E%/jcy^b&~f"},
	{A: "GBNI_gP^RlEgr3dXp1nlgPGJACQXl", B: "#@Fs^hNN1dgLRC_oxrdp394"},
	{A: "WUKa1OK8yfuEGP66T7d1x@N^-1~", B: "Z0zsf$zbmzMY$O^mZUCw6HogtkhaQ"},
	{A: "CmxUmO*EMtlT^SNZo$C#b", B: "%6_S$6JVfWmIDRZ5CED"},
	{A: "gR^_rEZkJ5", B: "a7##8J_#OHQYTC8lDrT4~Kcu5W"},
	{A: "aumCBrfMr*-HPZnve^z&nZR", B: "2l@YMzU/y4lMA&Y2Rg-$snw"},
	{A: "i&Aixy#tOZ&jj_DTX", B: "QyQcUm9F*4b*@t^or5ZonDJ"},
	{A: "WftFn1qqmHyOT2-SDKMQ*mPGA8TZ", B: "jzcr/7_W*_-M4a"},
	{A: "IOot8kOnrXiB&Fr9P~$Q$%C@%mU", B: "3gTQk5blq4%Xxh1_"},
	{A: "hjzi@%4s7$%8tW7dJ&5R%QOQMvY&~", B: "/RxBM9iO7cZ3GL%s6ur9-q$P%OG-E"},
	{A: "uVItmKk60vHmJLrtuX", B: "FqoGVL8eQO7jzYkP4vs4~29GBq&h"},
	{A: "DwV9d/ehbvXpiPn_pa", B: "48Sc2U8S68amd~uD^e19oJEZiY"},
}

func Test_Tst3_04(t *testing.T) {
	for _, v := range in {
		var state1, state2 State256_t
		var h1, h2 uint64
		state1.Reset()
		state2.Reset()
		for _, code := range []byte(v.A) {
			h1 = Mix(h1, state1.State(code, h1))
		}
		for _, code := range []byte(v.B) {
			h2 = Mix(h2, state2.State(code, h1))
		}

		if h1 == h2 || v.Debug {
			var s1, s2 State256_t
			var x1, x2 uint64
			var a1, a2 []uint64
			m1 := map[uint64]struct{}{}
			m2 := map[uint64]struct{}{}
			s1.Reset()
			s2.Reset()
			for _, code := range []byte(v.A) {
				s := s1.State(code, x1)
				x1 = Mix(x1, s)
				a1 = append(a1, s)
				m1[s] = struct{}{}
			}
			for _, code := range []byte(v.B) {
				s := s2.State(code, x2)
				x2 = Mix(x2, s)
				a2 = append(a2, s)
				m2[s] = struct{}{}
			}

			t.Logf("h1=%016X\tlen1=%v\ta1=%v\tb1=%v\tin1=%q", h1, len(v.A), state1.a, state1.b, v.A)
			t.Logf("h2=%016X\tlen2=%v\ta2=%v\tb2=%v\tin2=%q", h2, len(v.B), state2.a, state2.b, v.B)
			t.Logf("in1 = %v %v %v", len(m1), len(a1), a1)
			t.Logf("in2 = %v %v %v", len(m2), len(a2), a2)

			t.Logf("##### %v", h1 == h2)

			assert.Assert(t, v.Skip || v.Debug)
		}
	}
}

func Test_Tst3_05(t *testing.T) {
	// DD3DF87844DDCDA9
	in := "~AIR5QHoLa3ZM"

	var state State256_t
	var res uint64
	state.Reset()
	for _, code := range []byte(in) {
		res = Mix(res, state.State(code, res))
	}
	t.Logf("RES=%016X", res)
}
