//
// go test ./... -v -manual -count=1 -timeout=0
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
			hx = state.StateAdd(code)
		}
		conflict, value2, size := storage.Add(hx, string(value1))
		if conflict {
			collisions++
			fmt.Fprintf(fd, "%s\t%v\t%q\t%q\n", t.Name(), i, value2, value1)
			t.Errorf("%v collision=%v i=%v, hash=%016X, value1=%q, value2=%q\n", t.Name(), collisions, i, hx, value1, value2)
		}
		if i%1_000_000 == 0 {
			t.Logf("%v i=%v, collision=%v, repeat=%v, storage=%v, hash=%016X, in:=%q", t.Name(), i, collisions, repeat, size, hx, value1)
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
	{A: "xb4KVA/MBkUCGhPl^#V@a/", B: "DMYKFq-KS%u-"},
	{A: "pafmW3O_vs", B: "7eTd//Ilm%0nn"},
	{A: "H*&Rc~SDHf9", B: "mee74vJ@BoHSf4Xi/5LvC@h"},
	{A: "$gzwh4Pimv", B: "W^w8X^60Ev"},
	{A: "0rkvE53L@qJvk$yBy6xq*X8*@iz0", B: "X_K17v7kF@-%02u_MH8Zutyt"},
	{A: "De4j/6HlZJhqoZA", B: "qG3mHgh06Qw8W@&@&"},
	{A: "KKATO5mQH6X5aSFO4T06o", B: "E5Ras~Z6X4bh&$10xLxOaPApk_kL"},
	{A: "*D5ZZC/KtaCPxKRAy/b0CiM", B: "v%8HYFui0l32X9*RE~u4OJ1oqGc"},
	{A: "ImLDafQpM@S9cd/o4@", B: "A~AmkGA1T@Waaeu"},
	{A: "TqUZbJRM/sCaCN", B: "x@hdWwONwnal7uSDEOCp"},
	{A: "$I297qG^b6rmzRLs5n", B: "zjimiwEv6E&AQ9f@@rN"},
	{A: "qT/1f@CDUww-1", B: "CY81c%Y85jslqdob"},
	{A: "eyTgrxiK55wBOeBf58cwfb", B: "GxXHZmx0QVR"},
	{A: "lPitRx1fnv", B: "OmO/-8~9_v"},
	{A: "5@i527ldtj9", B: "LykBtgkjb_an"},
	{A: "5*K7lp9Vqt", B: "E#m&58*%Ad"},
	{A: "P9yfH2jyy3A", B: "3HIGqrjAj7A"},
	{A: "~nR0b3~P#cqIk*HNnXZvM0X4QHg*", B: "IAxjeFpnqm9Eu_oBMqV6n"},
	{A: "vNUgtHRAOa~O5y#71F-ZnT", B: "v4yP9/w_ai8i@qfXOg"},
	{A: "55VcUtK%k*WZhH", B: "k*J-W27%hA"},
	{A: "orhd3cJ3HYV3nhzJ", B: "#r*eDkQ3HKV*mp0%/W31b5jp"},
	{A: "S&hV7ryWeu1yv$GAl", B: "EFy/funA8i/JoY$o7Pogfb_zl_DCW"},
	{A: "~sCAdANJvme77WDCiF/&", B: "efU^88USk37Wc@oKc$#J29K-YQikV"},
	{A: "eb1U59b6d~qG", B: "r6&$0^@c0KXeZcAdPYctKAN#Fv"},
	{A: "q0r_4TKrzXnS~_~8", B: "8@X#PVFBC^e8R^BRQwTTHoJ6"},
	{A: "rO&t@PpC&qAIqfqSOdINaet", B: "gDoWwjyRsqbV_"},
	{A: "bTFU4^iHcqs1nk@BiBZ@4u_6Me", B: "yzpN&tWQCMgNeO$vpYG"},
	{A: "h9JD*fwMDDZ&FboV6O#5", B: "fLFMky1Gr~z$YB5QU9@f/mU$DoGD"},
	{A: "hO#dRFBt&jhaNZT~-2d_7K", B: "48EVT*bqz1fNI&eLb/k2Ksyo"},
	{A: "i02p8xMQC8Z1biogw1^~gSK", B: "cq8f7roZ5tiTdIrS08O7^J"},
	{A: "u_44Eim7hHhtbaojZXLIEeIO", B: "$cwVoWMtmMlpGa@Ei/I1sni"},
	{A: "ZXl9e-WXx4Xa9/ayN5pPCTG&", B: "PmAJ-T2ZrPUoa"},
	{A: "FjMSilsVx$xh1qra", B: "byiEV_urEaPdP*eBe"},
	{A: "vl7fQsHY1YuqOhTmvn8U//TWg", B: "s@YSvbxEmot^4paW^9eHEt"},
	{A: "k^GICsY7oV%ou$FAgb1&/eUSMAJN3", B: "d$2Dp#Z$6/4wAf6"},
	{A: "56^rAA%SBvDXwTLt2K", B: "UmC@AUGS6fK"},
	{A: "IA$pa5hPLRcyW9C", B: "0R5OIzy*8Iu"},
	{A: "&@2l2bl9D_XCj1JyyH8o46I$$Xg", B: "DB_mInEfwhX9HGv&WT"},
	{A: "2jes9_c$tl&17BMdhGlK8@TyrtIbN", B: "Vt~/9Yi@b*#jIX"},
	{A: "Go7wyFNoct1h1F$@jfiD%OPA", B: "QRU/vzg@Kb13Le"},
	{A: "p#@5Q5BQRF47MhT5V5lWJ6^q%", B: "mEWL9$^0n3Dq2"},
	{A: "iK9Snzcd5rQT6wK", B: "GclbpPTdpT7ren5t8jN*"},
	{A: "cw_Lisy3T3JtiisWo", B: "UT$Y-GNb@nAxV7u2Ql#cMSW"},
	{A: "B0tPSVRBnucdIKc", B: "TooY-/p/vmtyEl^0^Yk"},
	{A: "CWUu^R$RDWpnY--x&ZgHi/oYI2lD", B: "lC1T$%O2RzZZigfQPj8hWs"},
	{A: "P$V^OePNMV#lA4S", B: "F8~tSzeA-yj87/5fjS*JGaar"},
	{A: "jclA9G^qqfDFMEre29/N_a#0g", B: "8~IZig1DhQ/"},
	{A: "/^MAZFe*PcdPGY", B: "AgKIaSMEgcZw"},
	{A: "ak^@Q*sX^4~", B: "$9wspye~cKPg0tErrGLa^"},
	{A: "bfV#wNWQ^2dAjlN", B: "Uz%em2jq/2x"},
	{A: "kExll0^q7vW4eYD1~Qf0b", B: "lSK0EIx&H_~oU@Y5qnS2"},
	{A: "z0ojFWJo/Cs44WEy", B: "bBw1kk-Ex3aB8sU9~QG"},
	{A: "lD6VC10pTs", B: "^*DVyWI#y1FyLVcN2@"},
	{A: "X1DOviGmAZABIXGtc", B: "AKz*PKBTP_H5/RH&"},
	{A: "-XNHIjcsEmwJ*5MsHO5R2y4VJLi$i", B: "K2Ayrs-pKhQ~eFLDHoY2p"},
	{A: "dH--dxoCZ&Oh8R5lAP3", B: "O4L9S$dGVQu_/Lyd9vV"},
	{A: "78n1grA~pnoVTX4pnfX#wb@*Vvj", B: "wd5BeXLC6c-&I0mw~_OM_%Ppl"},
	{A: "/S#mPguPnRlpe*A7zYKPmrgQ*fGnb", B: "3duXlqY@J/1FgsLbh#YDtOsMhFN"},
	{A: "X^yE1^J$Ua^", B: "v~@E798RPADz"},
	{A: "ZM~ojDGeTIJlZ*dsSkcq_96xK", B: "db-D~K^kww_y^E"},
	{A: "d*#9DW9XPT", B: "#916UaORXrMamxD*#ZG@X"},
	{A: "XwIEMM$568A8YUL@gvETcUSgR", B: "%MeT1MU_uVMZr7J3b$Gdbm6aY3Xhk"},
	{A: "y0/N#1T@hDg-vLb", B: "qm_u#d#BRp5QF$8MOYjD7UnZwR"},
	{A: "Nr%7OsthcB&N~#YYerUr7uX", B: "aOF-fTglIJWaM7G^RBC3fYkvvsZ"},
	{A: "Yx0OpGNv#W8etbc1", B: "S/$l9Li%860#pLZ#_C~^"},
	{A: "gqO@FrS0bcq", B: "dLXaNPLek304c*%n4mgg9Z5F*9"},
	{A: "wsJ*sr6*GK$yl~C#wN~LvZM3k", B: "YG^7tbeq#MUNbt3~ixzs*&"},
	{A: "Pl$lWuTgVBjK*Cissx1", B: "3f9OTvO9sdZxQ3heBM"},
	{A: "ZM_Q~Ok7NXYhM$s9@", B: "Q~0S95UC@X^rbrCXY"},
	{A: "0/1vpkQr4CZ$9z859JP*YGW", B: "PrClun~zX4J^SO"},
	{A: "_kHjVu3lN7KhdHbyQ", B: "y*V_%0BBnAT7c#_2e~"},
	{A: "v/90C0T#6%~fLi9THecDG02", B: "t17kLva^pHFNUFcLVT"},
	{A: "Qx8GMeeJmi__7E32d*8gl6Z", B: "*CokMPvn~0"},
	{A: "WYW&PupxDGi7~WpbH", B: "TiAQT^TMM50KZ1g"},
	{A: "Ey$^plnW#Q/BZHo9Oy&PDfaE2f-", B: "mo^AFgUqgn~_#5GEjpk7d"},
	{A: "KD05700cig5BQowyrtEZ", B: "r0xZpQc&hS17auYsf"},
	{A: "q9xyHC/fX8g3h_T2p*w47/uf0lF-I", B: "FvHvBZcAupBB93Xb2tqgDY"},
	{A: "8^qUMRfpp%pxeLcM0d^*p8lH0", B: "GWbTG0ZVnE"},
	{A: "%GuH-SPtxqCFsVKG", B: "1KE%6t91ZZ*TpOliI$nS"},
	{A: "eVNvCtwjAd87h6V", B: "IMu_Sp_O5**sUgX7YYTB/uc2$dg"},
	{A: "NIoqp^osv3ZK$i2vYxgXdJj7/I", B: "1qeTRYGhKvZs"},
	{A: "AaAz3gZIlowVjIgdbmnAU0J/X", B: "o1Q**R5W-#xV8$"},
	{A: "tbdON#Hq85mBQ$IL-rK%w7q^N", B: "7K#p9tw^e4v6uZXC^DVMHserL"},
	{A: "_@EFPfTBtwp&J^hfdG", B: "sKsFuogvL/#"},
	{A: "tI3t5_E&5Q", B: "_7E51#W6_&LjpdbpKi4"},
	{A: "MSMps^NS1C", B: "IDlq31J@t5h%Ce"},
	{A: "tKUkdU9Po5", B: "bMUHdtfMoE"},
	{A: "mV&vLPdvN7bC2R7TZC", B: "p&y^sIhqSF41%/19nu51u"},
	{A: "zB^_W9#Fph7yv", B: "Lgcmsj84M~b8wGOy__"},
	{A: "AxPk2GCqQ2NfpulZ*y@9", B: "$ZS%M&Xqodk43ctr9"},
	{A: "wspEBB00FyGlxWRVo_UQ-&^hIk", B: "Ri$DP5Qln-mY~"},
	{A: "Q&2Ys__y@v*", B: "r9o2JCUojIIpX54/X_YX-F$It^Gy7"},
	{A: "_Yvd43IOBT8GN8yerI", B: "zNWE~bP8Fda0tu9dYO"},
	{A: "T_48s2g0RV", B: "xuJu4/$Cb_%-cgD83yATWAo"},
	{A: "08fzDN#2vSdW_UU", B: "$oKS^Iq@Sg@DFQzm*eKX48_ak$z"},
	{A: "uzcmP~huh8W-ojvvjQA_U6MK2o%UI", B: "zLu3uIOwG7BepRAh"},
	{A: "/ea^5Ib&62tboter^QXkOxuI", B: "P@KjUs1Fyv/vSCD8bcsU*EL"},
	{A: "5XE_/R*D8yHmQp", B: "UMLfpjrx41rD8bo"},
	{A: "AUms@aUJ*D7-3j&yS9Mrv2tthCS", B: "qjLcEyd_SBM8Y@~"},
	{A: "v-2f3$1iB-wxsAQ&E", B: "uWC8fj&7WOsMR*OZE"},
	{A: "Okgk5bYbZv$grAQ@gEd34RB", B: "LoSJh4fK&S0%YtXv-OY0hb"},
	{A: "Z-3fvZtBGJf/dWJGhK6j", B: "1aw6^cW4jgPs3~c6I/VO"},
	{A: "d8Jy&K3WvjvLS73xUQXe0XCq4PR", B: "QfOFlx6iWZi85#RTusN^UqLPQ7"},
	{A: "$7cF39_iGaP-MHoA_eZ^g@v9Y2", B: "Zd0FloD#aQyyWhKcEOd@gMDLOF2z"},
	{A: "j$rryaFOgjpyzE*^", B: "Y~clW&vg~PVPkjX&Ud"},
	{A: "nuFHg@^8%M%ioXhJ3OsnAXy0", B: "6&ldZHgy8WAEio"},
	{A: "PuLq5Axx3P~aoIcK5hjAgOqul~JQ", B: "Xe/GfKnFazK9EK"},
	{A: "c0aKr5ZEqjoQsQSk3V/Qj4QTd", B: "qf~1qYd7T^y~"},
	{A: "-b%Wy-yD3g40", B: "Re*z2rzOdtd5_M"},
	{A: "D%HjkRylvwr*BfCXfi$07eeQm@", B: "^nlc%vUO08T9A3eUd^a~ycC"},
	{A: "%dSL5PM&aDKF-#XV2-cjWZ~pJA", B: "cmp2FwW&gKJNQJ6RDL"},
	{A: "BdB-3i4R8Sv4w0j$XJpo88gZ", B: "UwJZPwm2F2%~wC%Yz*Hl37"},
	{A: "YsM&9C848GA@TokU4P314ees25", B: "LMg3Put_-PBuf93yDCOKbSEG"},
	{A: "ED5O@GRrsR95HIDZjxR5", B: "sRaR2h^UQMUcvcCmk"},
	{A: "TVg5DmIOaJ$wM", B: "G1butYKdq9tvhGpEDDwD_$lzM"},
	{A: "2RDHq0#g@PNDKIQzYL^er&", B: "A4gzMpH4G@bAP@mzb"},
	{A: "97AxeNvE%KCR1@0VddO8QA8t0/", B: "&^jm3qxoXy^u"},
	{A: "vMZK#IqHSehzve5f", B: "b&FAU1-H9x@@GaVuBZM0I2k"},
	{A: "kSzno4uqw^s4TuCx*E", B: "Rxu~xSkyMQwDV107"},
	{A: "xOfVgExdCb5^ljv", B: "lj@SUnxnr@*ANx4Z9#I$_"},
	{A: "Y~WqV1nF3wRw", B: "~SEYR_KILCr/"},
	{A: "cA-xluViheqkcok", B: "_z8J^P25DXCi-UZz"},
	{A: "S-^DJR%0hL2Fd5tuW", B: "IUc#Qy7z4f_~$Q^LdptBU0rMr$"},
	{A: "G73yeWsRxJPzWM80i&xI", B: "K9N%3/FTwdMo_BT"},
	{A: "%_8EWFaf6sIeMXy", B: "cDA~VvtbC%1UzMT1D_Na"},
	{A: "AqTXF6%6^n1JPb&p8i~", B: "/LU*Ctr3wSSL$OFf@n8gAhv"},
	{A: "_x02MhDBO/QYohG4a3rU$3tSkd0", B: "OxQKPVEQ3ssy0/k9R^"},
	{A: "R#cS2tkaGkrjI", B: "JluhuvR9Jq1"},
	{A: "5fjbC-dKuxNfpCK-gRyigI$uYIO", B: "$FlZ-tbdVVtSDsCmOM8vuqBIm3zX"},
	{A: "yEut3%z38$8rItGd-vgo/W", B: "QP7ZdG&%ryB@^4rvrNTP9wiDsPeIa"},
	{A: "iWDPopK3qhR@X4", B: "YEx*0NgH@smY-HYyGw/U4vG/"},
	{A: "_I@c7sCPRtTAjB3RTv_", B: "@B8Gt*s&TlKXqX3FFO"},
	{A: "&TPq1moev3GvLLYDP%y/7@h", B: "VX7/NJbu/4c@FlLIY_4O7~"},
	{A: "l1n$*aHh*E1oeT_@v", B: "Dvt7feoONr/fE1K~"},
	{A: "3#th1ioie_XF~o$4Tnqo&HhtB", B: "3B%yq~E-OEH4faQ"},
	{A: "J287iIdbDO2B-wPhL3m2&315$Et4", B: "D/ixvvM#ja6U~_uh*$KN"},
	{A: "VHUITwL&uGBaWgRdEns#$3", B: "_QUR*T/GDLcyX"},
	{A: "JkMtpgT5N9aD&%Jg9^*0I#2UMQy", B: "8n%0RT4cBzPZX9h#MlV6V&es"},
	{A: "Q7#6*ED0oEKCSV#c&RTYQkf~ieg", B: "dKAy2hTNB%/m2_bSKf45_Oe%j6~"},
	{A: "@%LX~1HmUxtlg", B: "G8ZZtfb/*q94~~#_QPp-"},
	{A: "2pYM6vg&p0Gt3J3FC1WA3J", B: "N-0ySpt-sX_2~zp"},
	{A: "q*xI1rZz2cozGV##0IEEXx", B: "5DV$_E-XU78Zb~i"},
	{A: "E8V1tnBbR-tt9mutCF&YlK", B: "AV3*Ir@XGn$7@P&$k%A0"},
	{A: "Iu_9mrbJceLdmm%b4sH4C#MM/", B: "fV%^/CFRYy6vaKAMNIBa"},
	{A: "wvEZso~SYDItH1", B: "iZr1f*4afWWIN5t8Ykw4"},
	{A: "OS4#m#OKMuBONx%", B: "QkNbt$~_4SPg%K6T&_5"},
	{A: "QoSCSacz3v7mbf90e&1", B: "XAs4bIgpX9L8"},
	{A: "qOjTOHHvLiHegNP9BLDZhS-", B: "jpvh%RH#CmpAbaAs*U8"},
	{A: "O&05h7NrA1~ZOpsCv%d-k2LDw2", B: "pGBnHvfyC3p#3_CWCO"},
	{A: "6&7z0$VdZpTEiNcgfs67q", B: "wzO%/Bn8YyY*3eTt"},
	{A: "N8aXn*mGdHhFwrN7pDERAzH6s_y7", B: "%UiLpiCuw$sVRdRqvgwydA*aGq"},
	{A: "oHK64I8uqg", B: "kn8n8qhw45#hJ*mY8XMQr"},
	{A: "b/VgT^GV^$VV^vIg&SiuBEx3", B: "w_EVRAW9~x@Ymoy8ee%sd&wA"},
	{A: "EpE_npjY14_OE10lzN/9lz", B: "OhfvRGXcxf3vT%Ia%uyN~"},
	{A: "4i~ivqXOzjUT", B: "9O##Kb9q5WUT"},
	{A: "&5v7TRoKy/EDN60VFRNVM", B: "wxkzr#0#aJg^GHtC&hC*fR6@jbX"},
	{A: "mi8Z05hl4/Ai", B: "#hmc3GTazP"},
	{A: "Kc1KP~Vf2#&N4YEW5ukX", B: "M-3o8Yh&20Df1JKD~DCnAV"},
	{A: "xq0&H@Gh~JA~", B: "OU&p_r1%AZWo#SIfS2ag~#_SOs3wP"},
	{A: "3GHBN^N/mGMBBrEo9bzIa8o", B: "qG&ef@Bc2L$YB729ps^iX^N"},
	{A: "@Mx4gu2yoxDomWB~", B: "owDPc72/M5_"},
	{A: "CIKqON_O9JZMQLYCC%eq", B: "s&PQ_q/V_hQ$qU"},
	{A: "#ZzOxKGMDmH", B: "oytaV69$G*PMwZJ"},
	{A: "jmffJt/JwFO2-sY8ORNU&0mqS^~G", B: "6bgZ3K2qttcU"},
	{A: "RFp5N2yES8%-52_0w5ow*nrN", B: "nT3yoTOEh&li1xoje4M_"},
	{A: "Q7wFn%WykfBpKrsNf", B: "_8G9GyE8c8zQ**d#eF5^rmZcRK"},
	{A: "t&cGP-Vj7/DRQPR", B: "93_nWxTXJ-BAacC$Wq-os"},
	{A: "eH-x^dglLS^Ud~5&a05IB88O", B: "g_KCqjWr@%"},
	{A: "DdA#2Ij3PrXuYT/JRzD", B: "QC&3T#OBEcZiPUE@iXvW%$x"},
	{A: "_ui/%yBG95J", B: "C&Vr&7MbIa~4OQ7OwN"},
	{A: "W4#Lce^KliFlY@i1&6CmS", B: "KG/ka0_ePc"},
	{A: "DT_0$3Tzw$h%Rj#L$TOxGvm", B: "_1%Oasfl%#WC3@c5SAr"},
	{A: "$%aQWqgz8ovp#Ha7/Fl68pAQfgF2R", B: "oJs7zTglVe&OBza1#"},
	{A: "1T6occ^IOy@1~I$/rck0T2l", B: "OR2Vr%Quxj1R5B8"},
	{A: "bqyQBjCIvn1ivToJ", B: "JiC%we/pnzRT_%AZ"},
	{A: "GGU&Ay8EW/U50", B: "ykz7ok%E3FQJ3dbckTl&%u3z"},
	{A: "tXyPx3C~swVi&aPo2x", B: "9/95sLB4IWw~ECIN@kmY"},
	{A: "KGCM9-uDJlFER%T_@w1*E7", B: "U0bVmFdDoOqEM~E1"},
	{A: "g5V561ICTZ-Y~icvVpj", B: "/8%*/obgDgmpebR#Tr"},
	{A: "UXGk6Uvo^^7ts", B: "P9469GSCCYl6TeLxGS_yym/6%p2"},
	{A: "pV/mCVpUGw*3Da", B: "gpP3na~b9IEY_tu3pG%E"},
	{A: "gU~1zK%&3wu", B: "9ShR1U-Pdby"},
	{A: "N3#-HS73jI#c8_AqXSLgq1oR6bkxc", B: "Z5/F3ePdGEV&v"},
	{A: "wPfcwlEwdxA6TOvIv^", B: "8R^S1BVL#UDUbpvffcxqdhWW#w3"},
	{A: "9gcHHEF0Qs7L-B&v&~*5Iniczklea", B: "ys7cy-dbb6WMC$i"},
	{A: "pq--%I7&tVKvqkh", B: "v*mwhk2h*/L9M"},
	{A: "zki@9&d#qQZ~XYBh@D$pD", B: "K~5&N7TNmbOe^/974$KxQ25T*-iMe"},
	{A: "4lhxIK&v#RcE~nG", B: "GhFut@lwB@fX*~d"},
	{A: "X~z5BHc_OGzEe_P2r", B: "koRYmtbx8v7j6znX-J"},
	{A: "3RjH/gumphow#jFE", B: "QpyFf4YyGMW@DPxkfEacj&H"},
	{A: "Y&^vol$%46-5PPzO-yRq5", B: "/KjVE2vXE6d&"},
	{A: "^%b$oEZAPg", B: "qi*WCj-9/D5/lQ6&aD13F"},
	{A: "Tg%tH#M3KVO52o*ksg6Xp5%ZDHQ9", B: "ngQ_~mh1wl%xBDF6EVq8k"},
	{A: "elxahEXxh#igpaTl$rIws@j5F", B: "N*t@*5uEwwFM89ReAQT_u~5Au_c"},
	{A: "4TcTB~6z8HLUStTQgU2", B: "n%vLrg%s~2*&jS8kR&Tn"},
	{A: "gQOwu-rb/Hcd01S_#j#LNHb", B: "^/5wrtauqv4m3yY-ACI"},
	{A: "6/OQh3$9t^Q@@Yh^#@ZEOjN77@w", B: "~3zwm%pGZ4*#UY*xHFKqVO"},
	{A: "jh&gb5^GgUz*uFPV%%xSSiS%", B: "vZ0dwL1rqjHg/h/c/TLvX6H@U^$P"},
	{A: "klWq4SitljPqFRN", B: "T5MmH_ij9sHrl*n%cq$sTwPtZ$q"},
	{A: "#%05K2-exXyl5QJo9P", B: "1uPjAt9weEW"},
	{A: "Q@^u49GvWkO4g%lux&1UsZ_", B: "nJZhbTw-oYjHiwPoRl&bw-kItA"},
	{A: "15^j7F$WeDKXXsEuSmZKW1SO/Jt-", B: "yTZWKD/BcUE/$IJZ7nVK"},
	{A: "EaiNZKbSCVhCR", B: "S_UaAFC&~1e6l1aiwzTdQ"},
	{A: "aazJk#UQBJ^NSttNjs%nMy*2T", B: "*FS/Sjnzp##"},
	{A: "Q_Jf5wQF7I@wI4%rE_G", B: "trCBw%_-sKql$~"},
	{A: "GzJP-uy1UDt", B: "~@qYFop9TdH2j1C"},
	{A: "eJ&q%bQualir@#v3y2ufSE", B: "3pfhfPPxcuU$CIKINI^nPlWdm"},
	{A: "O_B/s$avG0%PtU", B: "ss4AijohNazXfpLlzxaHEtUi"},
	{A: "oY9u1svy5FPJZ1fhyphxJ", B: "9tNQB5vW&oC8"},
	{A: "auDX5GD*An1t", B: "xUISSfUuNaMobjBMVDcVlmKb"},
	{A: "2%E3i2Q-s69Vgi_@6", B: "4cI9W%l8V%ZOx8qvdgwD/JX"},
	{A: "VQYkC*sc/4W_r", B: "2vkDt&QzyV"},
	{A: "cs5VHHhdKelxDKdBCuUca@-oj", B: "Jn$9tavffLAh7Bdd@I6B*oC9HNi"},
	{A: "5IgU7DuClXTE$8y1PKTHQgpQE", B: "HvwPB3R~2a&VD4Ay*%qNoYi0"},
	{A: "SW_AgTre#s", B: "_%-wJ8KCR/lKB#"},
	{A: "24mfxrNQgIVbO$NDTaf", B: "xIXVB^2OkT8n9@5s1GjU"},
	{A: "cJJSAUI0~#b4Re_#L4lGn5", B: "SuHJMot71GK-1CVTyFfc5lCzCP"},
	{A: "^BI4GxvUo*", B: "^Bh4fxvUZB"},
	{A: "@3QUXu1A&mC_p3cNQ#QLNiM", B: "#iC0@ANNniTi/2afnca#7"},
	{A: "A#R5eM0rSAAl3&y", B: "~_/BmueT7F7Pk&4gPlx%"},
	{A: "B$S-_fWt4A0/_LOGMp*n%^%", B: "IcQDAWT9P26*-gakGNJnij"},
	{A: "s31iuyXKYoKoyBR12mLcs", B: "E$0/~3oD&ck75vTxZCbZ053SoP#mH"},
	{A: "Dp4#$0Lcbi7JG%~", B: "SWMkilRLaf3oG^p-rmrz6$GU~"},
	{A: "@6TWe04uY5172", B: "TB*/%QFXX-#t"},
	{A: "wCJ7V_idxk^X7hFXy/zMJ0qSJ5Dt/", B: "#Nn&gBDQEf*D_8_LAGXrrmBq7m^"},
	{A: "P42G^ZJD8CbqsmsOe_-QSq$Ya/Vjx", B: "nxn*e88vVB@7x"},
	{A: "%bFBrraG0AVJ0wuq/pt7u", B: "v5j&@P5JqbwuzrNpszv^M"},
	{A: "Ij#2@YZjMx4_bg3kJ~bBuR_y", B: "RX2M%Y%NXdd~tZrzm01OD1"},
	{A: "J9lToGW$gl7-", B: "%ggjvn6HQln1PQK@"},
	{A: "tkBSf5AurL~p8%XW", B: "Y_#v8%qtMX&7Q~$Hx41"},
	{A: "Rb6w_Rqr7WPGf$biPIiON", B: "/UMRjMdfjOf9~PCXpcAZ&GoSna5R"},
	{A: "#m%ID&4lxB$6OLI%A2FA#Z", B: "g$dh%jcajAbdU&KMAZ"},
	{A: "AAAspkj^PP6aKwz6yN&MEpd6i", B: "kaD~p3iIwxVHmp"},
	{A: "$@gi*24i*1Ipz5s~2Fe", B: "qJL9idlwAOSEsMqCHvEP%"},
	{A: "qK@$K_sP4$raNBCD@JsgLgsiN", B: "8LPZC$&~1wQ@z5grc@blb$qpCQ"},
	{A: "ziC&j6gB1zsf9Mb5wYPQ#kU", B: "4ru7$/&6^U4-kwybF37hPIeu"},
	{A: "$Pcrk#O&6ya95IMiT2J", B: "JxXX*iluEcyc&CVpy3&w%fnP_Kcy*"},
	{A: "mtgrSDS5cmgcqqJ-d", B: "HtB0A579_4Ocw/2*3ONUSJ1PFpNQ"},
	{A: "ycFdbXf%gL-4AZMT2", B: "l6KRwZkNro/5N3Ohf#tduh@q"},
	{A: "W9wgHBwK9Q", B: "NvRjpqIsF03BIJUktHlArO_dU7"},
	{A: "i92nEChTC4OauK", B: "WjJ1RRvnre/viH4-K"},
	{A: "-j7cs3^/peKahBMKfJubP_0G", B: "t1Sv&K@a_GS-pi@#NzFKs#"},
	{A: "lBvMW/Cdy6yYwC", B: "2GLC@WIvzb1OOH0b^*h~D"},
	{A: "ukQpp77RdkI@^_@Gc~B&9/4YpRF", B: "Z~&Mks7@GHwzSSXgCX^EpSo5CeXvh"},
	{A: "wRs~E~/&-PjOoyKm", B: "4^B6aT9f7aA6c"},
	{A: "4nU56eosIZ5V26mSH^J", B: "PVg8enOSI^Eeu"},
	{A: "y2igy-SsGR", B: "SO#z0$kEFiiOj"},
	{A: "wLflyOwDiG6_JB@HQs8m~_8iS*", B: "9N%2Q4_CAyG2tXbP8S"},
	{A: "/R$PSxsnelVOsdHZzMJ_ES31XbM3", B: "YHjMc6fnpCOC2M@9Kql8"},
	{A: "I*gzH$y/#GlEShp2~Y3", B: "$P5PZbBw^GQZ~up"},
	{A: "#gDH&nCJK-QN1DwkhxUC*R-", B: "Z&YE2O69W3E"},
	{A: "LGcROXJji&zNGC^U~NzH0Ggo@P2", B: "P3hs_siYO#LrA2aYE"},
	{A: "Mo@&sgTR$i/qBgscRs", B: "p#/@NVVzq5@/WAE8PG1*7i8vx-~w"},
	{A: "BqwsOOW%L1", B: "2yEGwW4JfC1&BXnM@faaA#jg0#Nvl"},
	{A: "29~fPZKVg8nmZNoALA/", B: "9vORoq8jVP*EP5A~%d^^aWHp14Gm"},
	{A: "XvQa^6#pd&Yjo2@*GVYLTVZN&uuxQ", B: "#fy21$6I3/32dBs@ty"},
	{A: "D_T0ZiI_0SW", B: "9J~7rqIkhasC2ZHmkhrl"},
	{A: "D_T0ZiI_0SW", B: "9J~7rqIkhasC2ZHmkhrl"},
	{A: "&qLvL%Pf&Hcc_pyzr$*tjDfUrKTHA", B: "l/f3@IPdGWl381C~YtF"},
	{A: "/3XToPwoF^$ft8r", B: "/Lkj16H#6-aBfkfC4~bg"},
	{A: "D_T0ZiI_0SW", B: "9J~7rqIkhasC2ZHmkhrl"},
	{A: "&qLvL%Pf&Hcc_pyzr$*tjDfUrKTHA", B: "l/f3@IPdGWl381C~YtF"},
	{A: "/3XToPwoF^$ft8r", B: "/Lkj16H#6-aBfkfC4~bg"},
	{A: "m5MepA4ieI-", B: "d3u7e$L/7x&7QxQ%%Z5oTvGQ"},
	{A: "1PZYJrm&*$W&Gu8$0GF4zG5*HR6YZ", B: "__$1^6&-faAW0z"},
	{A: "K7jtIdT7EP1L1a_-XymwNn", B: "QKGz8BcZ8syFBSjwGICM7$rT"},
	{A: "vmle6u2N1jLwH@$Mq2~xtZV0O7#", B: "oZ_6mzkbi7mB6ixa~SgnXYCi^FVU"},
	{A: "6xuk8G8b0s5h", B: "~zgwOw6F9/Frd"},
	{A: "#CrvElITyZY@DSjkFPCTX", B: "LcdmbDsHfupnw^FCTHQVKeSQ_"},
	{A: "upnHtCk9F9Ju$tCQW2X", B: "SaHk0j#bofvYQlIg8"},
	{A: "CtJGJOWZL8wv&c@7/ZLB5^", B: "VK9bcaR/rh^"},
	{A: "2ZDFBQ45SlTny", B: "HAz0yeObb*IXXkMw5IDuFSr"},
	{A: "yI1EOKPBc6U", B: "enFJz7kS4A8"},
	{A: "hw3rJt1aImVAUCM8Dqtlg@SrtcT", B: "@N#_-lRxVUu"},
	{A: "O%_g2Iyy5J/Ze@jtW_a#NNs#bk", B: "FN5u0P#~^x*DwdPPg3_/PuR8r"},
	{A: "w6CH38#6xa", B: "-^SSAGRJ7K@oqM$lOlp@^Ta/VQNu"},
	{A: "0rqF^9GYeF%1q8iRWm", B: "IspxgRsHZV"},
	{A: "$&~BONXSEXd", B: "daTp~vVZs9#77Q%z5p"},
	{A: "lQBZ3Jm1@&2d7^oxJeB9bU5YcKV", B: "C-ykf3QC9OwWfzK"},
	{A: "QooDZ-$oYG", B: "jiq/nhsIptkiw-V&$8KP7_OwA"},
	{A: "5YFOx80-#1snaCLs-0S", B: "/s0ihpllKN7$TxOLB^"},
	{A: "iQ5iwdp7-n", B: "T7sIs%Kk-1g4@9w*"},
	{A: "u24dhi^o56", B: "E*yzZ5ni%z7E~J*KoLcl~~dm7rV"},
	{A: "CbqiJM&vi-894%NRq%WDxe*c", B: "K4g&Y$9UGM*NbjSkD5S7"},
	{A: "Tm4iT-m2Qmn2y80bPZE2", B: "Hmje*$q#_a8E_u7rP77C"},
	{A: "8D^1P&881/sHcXfb_RbVnz4V", B: "KG83jqeiO@57KSr1@9Js@Kc"},
	{A: "Y#LmiyGWE9ud%fB", B: "Khj&Hy/4rV/-"},
	{A: "IC$@MOcxYhhpaIyU0", B: "%-1GiRLy%n&TuiI%2yzi^5O"},
	{A: "8oo^uIEydNVJYV9z@5ZQ", B: "pqGFgyhHgjuOD4*"},
	{A: "ob%JTXldUhI", B: "00L%$&d6$XMf0gEph"},
	{A: "/pIyb/c8zy6yfwVSfojNUE~", B: "rJoya_3_GQdOA-mdcm5"},
	{A: "Hh%jjj#v-2woADs%yIx0eXQE_", B: "&tYbN7j~2HeoaqP6VnWrjqhxmUJGf"},
	{A: "n0-hE2Mzy#N9%fth4~P0bh-Sd", B: "tisKUwuv20i6"},
	{A: "iFN-fR1-l8YSYlgUQ", B: "O9Vr*BYcgFjVX-~_HYf~DFz11l"},
	{A: "$IY7BN#^d7Xk6dWUjcHekC##6Fc", B: "57B4$59l466o%o5gj/K%*xKy@s9g9"},
	{A: "qmc/lOIG4E@BJ", B: "^jpX$nxkKS1Q5rDDJ*ehRts"},
	{A: "5Kx*Bsfz5ejKCJvG2_1Qv", B: "%1/CVNC6tr_@PV$zd85r"},
	{A: "l^o$fO9mPD", B: "m^o$fN9mQD"}, // bad or
	{A: "S2oieLGyks", B: "S3ohdMFxjr"}, // bad or
	{A: "r*_z/U1NG4#IzzVw44t", B: "zz*UvOLI^gv0j$T"},
	{A: "&trAIU-lbXW2Wl#Ma", B: "45v~/#1467GpT/TaTGj"},
	{A: "D*MI6kDqvDY4U7qyWPlZ0", B: "pMXtLXswmgh-p%VffpTCU"},
	{A: "X2&cip~Wvlck7", B: "%JuXDhNn#tmi_Q9YB-HmeI"},
	{A: "Gu~cSXAoaOrxt", B: "Fs4#U^ShONLwd"},
	{A: "^2SLQNGni6c#5r", B: "uU7jp@A_eVpW3FA4"},
	{A: "HI*Xe8lbdWgtYEi&^Dq*_bO^bxc", B: "_sgPwBBYgL&&Lk0umk422wS*-or6"},
	{A: "PS#eSaU*h%R*Rh5tY0wulNpl5", B: "CsKw5$vEDqzjoOP$Q"},
	{A: "Sxbo7ukyQ%", B: "k@~Rz$^/-~$WK@v1g4Jk"},
	{A: "TPoQ5nx-OHaDV2P", B: "lcCgyauHd2D#xCEzSafL_R_C"},
	{A: "Vzs#njOFJ_tspa@g1rjie", B: "4wP^vT35R8KZ-/S"},
	{A: "bcHu&rQED7X~ge1v%1KJq&K-QK", B: "ViIKlk8JE&VYNQuBO#PMbeQbot"},
	{A: "kg^Q0mk$4wbd/*", B: "atCWUS4GXS-hBzSb8u7N"},
	{A: "247UpK0gi&nH-0L6wy1Ad", B: "v^frKeIunVs^3l/KWfYTw@zk^oX"},
	{A: "FPUMHr#F_YnlBt^K6%zg6JZ6", B: "93#lqfGSW-uCjkA3-%@"},
	{A: "xW-X*sk^N6EKT", B: "QNKMeDjBI^5P&g_TB"},
	{A: "S#&^Jd0xhT", B: "GKhPace$dw^M&teqSx%CJfUFR"},
	{A: "CcqnVkeHF52JhK55S7q", B: "YI8LYT33cL9V_@"},
	{A: "L2HdWshs-zA9xTVTnm3P5S", B: "*Y~A2G~%BdgE/iF~"},
	{A: "_Ojd1t-GsrF", B: "zud7v-_uxG3gAPgLczaEXi%$hBK9q"},
	{A: "Dzu/F7r8u@8FuC", B: "k#NMLLkS-YK"},
	{A: "gtW0t0YoBj~", B: "Cp&pSz%Jck#UBg4XSRaxu5"},
	{A: "6o9&WEQp$J_M", B: "-Apm0KWG9sjW"},
	{A: "i^8wrkam^6*6~/mIyobcnfOVN&", B: "EOolCx8S8KXwVriFUmh#od"},
	{A: "L7RSD~2T4v8anAt_*tsVBFi-B-", B: "KERZ$~%-@n/v^Te_-%d$MKB7~C"},
	{A: "cndM1@RRQ8$w6%esCsNPWJy~1h#B", B: "UaaBwSEhMTW_H&cr693wKKQxz4S"},
	{A: "Y_YIMb8hx3K/~GlDB-", B: "QC4whmg_6X8-Vxp~T0x4B/ol@muq"},
	{A: "_ARYFiyq1-XD#D", B: "Whu^#w9ejdTYGUK5"},
	{A: "Ub7G4vNaDX0Zt4Zqa^1*", B: "vNidMrHGU2b*"},
	{A: "wh&E5wBA5Eaa", B: "ETz~X07Y$BL$4tgL*"},
	{A: "Yj6AsUxu@P8Kihq^yXW4s", B: "K_m6s5Cb*6zt*R%eQJ9"},
	{A: "ye$j^#r2DX^EAo~1KH", B: "aDrkAYUTyB"},
	{A: "-$vAW/UlCAX8WW54-hY7h5%gEm", B: "Up$Odc@GRd4Op4oMX-qin#w8AaI"},
	{A: "idPO^zrCcvCv-#Q", B: "4XVLRr@tUjJrT-iA5fG@8&$#io3eu"},
	{A: "fu@skSTUKMUuIb%ewluruq", B: "K0W42F$qxHq"},
	{A: "6MsGgN^oama5kFIVQ/-jGGAEo", B: "*/M%tR8jUukV*2zsc1Y_y6"},
	{A: "f_NxBJ#gGa", B: "Tr6l1%BT_#Ft_-8dT_1gQ_i6xdO"},
	{A: "La^JAt2kUz@Nl&srp3aywcCpaRE", B: "sR//nNjt8meZf_EV~T"},
	{A: "8drocX1PILMep", B: "klgwMRXd3hWy4fr1TdqLf5Ps"},
	{A: "qgH3@H1_YxF@", B: "jcA0Do&6Ba/_ol2aqK~OXH4"},
	{A: "geewFoKSmvML~ZP1r$sS", B: "LljdTKNqkMe~rfJ#w"},
	{A: "%y58p&hlfzHDAelYd", B: "MN5khPseex"},
	{A: "oCQPhEuLtk6jnwVTaW_33v*", B: "2u%jTjJ%zSlP1eq*#5F9oi5%AREE$"},
	{A: "w54HIG1XzxDw9a#p", B: "GgkKxBy0EQnLXM99"},
	{A: "AaUSOw&TE7R/fMd/1rUwXs9i", B: "cbbbVJ2Z*6@OwvW3~M5czh9RxZ"},
	{A: "Lz&bT90y3G44X67%BazAw", B: "ewOERkom#C*/P2*hmj"},
	{A: "S0pDRI_lnMz&L4Wqh-", B: "pPnJEC1EjnrV"},
	{A: "hKmWuv8$YZLro4b@y", B: "JJMuUSBxOSS0kv7iCTM*vH^rL/T5s"},
	{A: "zlIa8lBQ5Tl%Hg_5@%gSjN8hQB", B: "Y&GN0iKpf$~r9Ao41%Udv"},
	{A: "tXdNn2XUGm_S2_O/", B: "A~k%SaO%eMX9jJz85e#RoCM2@id~9"},
	{A: "9$x&fxsUN#c#zfdYRWfBQ#l^", B: "~kHN/^e&#Z5v$mB$3d"},
	{A: "Mch4AfFWvwmYQiMwGo6c#G", B: "uW#iVXu@0wO0EWAyzx7QR1BefaDar"},
	{A: "GJd1Jo@KOyFUT&3H#cs%$DpS%", B: "j7o8bkzVuh8q9OD13u5RM"},
	{A: "-oq-q5Tu3BU#^01bo#%mQL", B: "_KC~TUlrK/R0@AEHqjcn8qx371"},
	{A: "RnBOl9a187x_HnOpWILc^~hzaFxge", B: "3453Dl2W1rpk$URv5l3q@Dzo&"},
	{A: "Y/m5n_dkwi_3sbNm4Xi3", B: "eF%Yh-Ld8Mow^9MjzB0"},
	{A: "KfXXfwYnGg", B: "qXOq1DFJfnlc#wX"},
	{A: "li/*pfRH2j$ROrkh@QpqjerZp", B: "vH#ju21~FgDM$@"},
	{A: "AsJc4fFW1KTJ", B: "GG^m62EOV2FD1#/f8-ZPyF@av9"},
	{A: "QFnO_y%lCPV@RAl3kteO&2n89zfh", B: "91CiD^m8Yk"},
	{A: "suTxJCN^w#eMA~", B: "6-5d-8ovBSRkZw@3~G2C#kGqxz"},
	{A: "o2K4tjh#I#&*L#", B: "6e7OiNDx#aoea"},
	{A: "TP$ERWpVn25*0gr@@", B: "Rc1c5@RuL6DQrg"},
	{A: "Cmet%2cRfci&/F2qZw", B: "ifo92l57Ro$^LtFhN4GvKMb%DLAa@"},
	{A: "Ze$i4q0ifCQ93B", B: "iQfWUqWO5QruXY9E"},
	{A: "ovDhcM9dyXtvt%OJE3", B: "GabXej8PoT"},
	{A: "aZ1ov42YE4t-hNc9HG&uP", B: "D^olmhpYaj6T$P&&oPVJxusoya"},
	{A: "iMCAA4$ka_sYc", B: "I9v/#QxQt3r"},
	{A: "SGh&aXrH&2D3", B: "^L&qWoBMbb@7Dr3Yf2@Lss"},
	{A: "Ix*BYWno*xX6pRF3WpU@IKz-", B: "MaeWshIv002iBCHZqSdCw^oM"},
	{A: "H~Rz~bFn6W/&KT*ojVH^ijE7", B: "d%@*5a9W0D6wNU9^"},
	{A: "-Q5erjE/mK#UO5sG%1zbLW@~", B: "mo2hcRD4fABZ@"},
	{A: "RW_oS85OOGa371aRyNy*LDi@$G7J", B: "gIcM-shWUq*IbPSG6ol_VqaBf$"},
	{A: "zGQyd9I9cmSE", B: "*E*Ogl^lTh&8VY_U1yROT&"},
	{A: "KDzfY8nvnda&SHlJ3Tn0etM42", B: "g3H~3DrXoY~JNgz9~uPG/^_jQu"},
	{A: "KyJ9M$a9H5D840ae~mbvsc2J", B: "e3yOQtBbGC"},
	{A: "@/xMB*Ffx8", B: "hoDx1_Iuz*~A5xnmH^/e&B9%dw"},
	{A: "Q-e0kNFBwEfuI_/^_HG5Oy", B: "Vz7AhQYW@LB4d^JU"},
	{A: "ltf%65URsHjM2F/%", B: "J~HTfSSKYnuPXS*jj%i#mRMW"},
	{A: "I^XU@yPO@lYhJ&DQJek9ebt&@hh", B: "q$/~EgbEp2nOCfNB0uFcyQ"},
	{A: "OpT&V#nuU4o~2a8iQojv0u", B: "1k4$anYnj^VZGX0Xh_2/q8raj"},
	{A: "eq60Ox^uWm5wd$0Lnj%s", B: "Lte$biF-*C2IoOepCf3iib"},
	{A: "_ZBgf**RnPOjhnWdMStl3f#0cX", B: "N0E-01bZHls5ML06qU"},
	{A: "_9OUET_y$o4Rk2AG", B: "cm6-J*&QYMcg%kzza5AVL_RV/hvKu"},
	{A: "@EzOm~axvhgnft5p0-", B: "2v4wsZyf_&wB6LnwdAD9"},
	{A: "XQ%FB~K^SlTmmNSQk0Gv8IhDbC97", B: "1J_GNr8YuWQi"},
	{A: "LLZ2rp*RE214", B: "xDr^Du-hCc#-ch*7XhnieNWr45HyX"},
	{A: "&bZjaFftShPfyk", B: "Aw_$fOvrQt8SrYmiebKyEhFG"},
	{A: "fpdE*fp_GA", B: "wo_-ki1JjIY4ruU93G5"},
	{A: "oFfu2clLkC*1L#~4Z76SF95Qx2m0", B: "T/YmzTyJngS4SO/~/cSBp-00EQ"},
	{A: "Dc0h%Qm/40-AX7R*WDerbFF#Jvi4j", B: "Kw5Nmd_&QBOhgXfAjhO"},
	{A: "bTHk^dUExHBkEAxC", B: "-RCC%Ct2#z-cqPnbR_%f9"},
	{A: "%*CR#ii2EJ*$65N%", B: "TGOxyZDMRx~$_&Uo"},
	{A: "auwgz/^9AfGU6F$DoBulxODqZDiD", B: "4L8Nx*9cBSYNojBbx/I"},
	{A: "E*wksm3UKBn&-Xkra/4", B: "/m_@m/yGy2*"},
	{A: "TbWzHWHl&mc$gHyr@BT&C", B: "^^3Tpx&HyfaT1EK7@RG#~$u"},
	{A: "nQNAGY$foIZKwCX3daV/", B: "j~5QxM%Z%lOAGwI$eyK"},
	{A: "C/WMb^8qFKUlv6#-k*RiiKHzXYfo", B: "0Jx&i-%yPoDJJTmOYc200Jd^"},
	{A: "TUY2REs*0&vrR", B: "Wo11V~XAW%Iba3byTY5TvA0m^oe"},
	{A: "xOHOi6kMMWrL_8jq", B: "1lts^cu7@TmA22Hg"},
	{A: "CKrzwZQ^wuZ3krCsYgcq^t", B: "e^YasO0U6Osl*"},
	{A: "-icIsut@&wEF$/h&-DsEaIvWlk", B: "n2_B9MtvsQTmRlv7"},
	{A: "ZezgPS3gK8EX", B: "@Z%KG5TG$Ke5%^a-eK@l#xu0P"},
	{A: "uK$0^3RYDV64AG6Msn09~k6", B: "6e~rc^U7UY@-3b"},
	{A: "f$mC6-@_laRg", B: "qhXHvSjQ@GZchRJN98x33hY^"},
	{A: "6K7xzc_mdH8@ku$P#VVTQuZLFQ", B: "y2ewuqyGsI&V*HC_1P%q_CBk"},
	{A: "rr72_5FKECmM&6~rM", B: "V#7RkD@cyH"},
	{A: "vS8KK$r4E0Xq", B: "kZWi%G6T^OriCj_Zc~F-RJJ"},
	{A: "inKe*9g2FwkRHiS#6eov", B: "%RkcHjd%mqyxy&7pOT"},
	{A: "2&TNIXEThLFU", B: "W5Gjy%%VlEKEs8D2jq4MBC9"},
	{A: "Kz%CrrF0g0QGW3", B: "xNe67xFWjc#n_19#2Q1KPm~"},
	{A: "oPkA0eoUTT6^VTm1dGd$s#qKJo", B: "9SimjJnaf8PGHRbRRzrv"},
	{A: "2oJribnwKM*J_hQx", B: "FmfoflZ81~@"},
	{A: "glBqqkbNrkQvjP%pQwByc_", B: "MwhpMhiB7bUWOX6"},
	{A: "g48Aj%uTZwKFkrbv&ADO%$pgI*RUp", B: "ynz5Hoj_cBgUYyp"},
	{A: "imGHdKXYGKZtVB0LLZtIW7K", B: "v^hs65vu49vswe6_Y-"},
	{A: "LuC#JKPgtW3&5H_JiCE^1", B: "PUmfCZDLImz2MPr"},
	{A: "5Evsbh9fzwb*YAC7r7", B: "U3O3Wuf~Y1@"},
	{A: "v*1nlz7^U8r-%", B: "@CQkTxEA79L"},
	{A: "TPtoL^cgCNmKmS", B: "$KQ%n9zx&zF"},
	{A: "oKEi5O^gKuit^Y2l--/w", B: "ak_s1OzeZXWzd@3TN6l~nHp%"},
	{A: "qwheZA&WX&Z1vI*#", B: "_tB8uAb-7m@80D5g4St~Rl-ANjjFO"},
	{A: "/Eua46b2rdq&", B: "-sPk5EJ/WXJW_e&S"},
	{A: "YThHg&i#8MazqOcnmN/A4YZ", B: "bL598GXN7oL-ws*_*"},
	{A: "ZzGCarevJ6&osrkN^-z", B: "Bpxy4cp^Yo0_6FOR5D"},
	{A: "vJaQSYDo%ZS59x/A2vSpPc$", B: "shl4&YOZTct*d1nRsnv$#h"},
}

type Res_t struct {
	h       uint64
	a, b    uint64
	state_a uint64
	state_b uint64
}

func GetByIndex(in []Res_t, i int) string {
	if len(in) > i {
		return fmt.Sprintf("%04X %04X %016X %016X %016X", in[i].a, in[i].b, in[i].state_a, in[i].state_b, in[i].h)
	}
	return ""
}

func Test_Tst3_04(t *testing.T) {
	for _, v := range in {
		var state1, state2 State256_t

		state1.Reset()
		state2.Reset()

		for _, code := range []byte(v.A) {
			state1.StateAdd(code)
		}
		for _, code := range []byte(v.B) {
			state2.StateAdd(code)
		}

		h1, h2 := state1.Sum64(), state2.Sum64()

		if h1 == h2 || v.Debug {
			var s1, s2 State256_t
			var r1, r2 Res_t
			var a1, a2 []Res_t
			m1 := map[uint64]struct{}{}
			m2 := map[uint64]struct{}{}
			s1.Reset()
			s2.Reset()
			for _, code := range []byte(v.A) {
				r1.h = s1.StateAdd(code)
				r1.state_a, r1.state_b, r1.a, r1.b = s1.state[s1.a], s1.state[s1.b], s1.a, s1.b
				a1 = append(a1, r1)
				m1[r1.h] = struct{}{}
			}
			for _, code := range []byte(v.B) {
				r2.h = s2.StateAdd(code)
				r2.state_a, r2.state_b, r2.a, r2.b = s2.state[s2.a], s2.state[s2.b], s2.a, s2.b
				a2 = append(a2, r2)
				m2[r2.h] = struct{}{}
			}

			t.Logf("COLLISION: %v, DEBUG: %v", h1 == h2, v.Debug)
			t.Logf("h1=%016X\tlen1=%v\ta1=%v\tb1=%v\tin1=%q", h1, len(v.A), state1.a, state1.b, v.A)
			t.Logf("h2=%016X\tlen2=%v\ta2=%v\tb2=%v\tin2=%q", h2, len(v.B), state2.a, state2.b, v.B)

			var my_max int
			if len(a1) > len(a2) {
				my_max = len(a1)
			} else {
				my_max = len(a2)
			}
			for i := 0; i < my_max; i++ {
				t.Logf("%02d %60s %60s", i, GetByIndex(a1, i), GetByIndex(a2, i))
			}

			assert.Assert(t, v.Skip || v.Debug)
		}
	}
}

func Test_Tst3_05(t *testing.T) {
	var expected uint64 = 0x2D879748C924EE9B
	in := "WrDzIBBp605NfFUP3%Oy8dh@"

	var state State256_t
	var res uint64
	state.Reset()
	for i, code := range []byte(in) {
		res = state.StateAdd(code)
		// t.Logf("%02d %016X", i, res)
		_ = i
	}
	t.Logf("IN=%q, OUT=%016X, EXPECTED=%016X %v", in, res, expected, res == expected)
}

// go test -v -count=1 -run Test_Tst3_06
func Test_Tst3_06(t *testing.T) {
	var self State256_t
	self.Reset()
	for k := 0; k < 10; k++ {
		m1 := map[uint64]struct{}{}
		for i := 0; i < 256; i++ {
			self.a = (self.a + 1) % 256
			self.b = (self.a + 1 + (self.state[self.a]+self.state[self.b])%(256-self.a)) % 256
			// self.b = (self.a + 1 + (255)%(256-self.a)) % 256
			// assert.Assert(t, self.a != self.b)
			m1[self.state[self.b]] = struct{}{}
			self.state[self.a], self.state[self.b] = self.state[self.b], self.state[self.a]
		}
		assert.Assert(t, len(m1) == 255, len(m1))
		// t.Logf("m1: %v %v\n", k, len(m1))
	}
}
