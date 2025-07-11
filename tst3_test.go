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
		buf := GenerateString(rnd, 10+rnd.IntN(20), CHARSET)
		var state State256_t
		state.Reset()
		for _, code := range buf {
			state.StateNext(code)
		}
		hx := state.Sum64()
		conflict, temp, size := storage.Add(hx, string(buf))
		if conflict {
			collisions++
			fmt.Fprintf(fd, "%s\t%v\t%q\t%q\n", t.Name(), i, temp, buf)
			t.Errorf("%v collision=%v i=%v, hash=%0X, storage=%q, buf=%q\n", t.Name(), collisions, i, hx, temp, buf)
		}
		if i%1_000_000 == 0 {
			t.Logf("%v i=%v, collision=%v, repeat=%v, hash=%0X, storage=%v, buf=%q", t.Name(), i, collisions, repeat, hx, size, buf)
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
	{A: "wc69qt5aD#m", B: "IgnB/zt0tIaY4y5WO"},
	{A: "cYp4W@ARg-v-ZE6zO", B: "Mn2s~HnQ~^2D9vAOCeUg&6EdG"},
	{A: "oJR%$L-ojMwDKLQRsgTZ%Ld4zJ", B: "RHqwScU9Bb/hAGS"},
	{A: "ieFOTMDRf26KDXp7X89HtKMZ", B: "TB@Fgg4kwvZ4qcj*e-k%tGW/o#v"},
	{A: "ZW@5A@PNeaVgWjxilsDg&", B: "~B-ApzTs$wL%"},
	{A: "yUR6QTdn*$-6-Mg", B: "kc^F^#trYSQpL2@YxrzsUG&kDa"},
	{A: "/*&IC&scY3lYVXh_s", B: "c~%qYSSCWVwyO*d2KMk2mzEL2x9*E"},
	{A: "bYK#ou-%/_lowdTJDtqk^FT&_dPa", B: "D/Z9OVZmQ#EQyBBC*Tl"},
	{A: "BKytNPwqzlEXB3C6ot*th#", B: "_qUwrmVY4f%wmqj"},
	{A: "24vWqxCezX76HWn^160", B: "/%UWIZSp@am^NG"},
	{A: "gIDFraDf&Xia-U@2&e", B: "5rkzBp53C_TkoD#r&jEXYhkuWR"},
	{A: "XZtGF2$nW3YMqHpsTlo%PVj$", B: "Dmf85*ulAE#zMFdZCn9XQj/E4F"},
	{A: "u9saH51UD_j", B: "4T^piK~7S#*Ah$x$sHFiEPBd8"},
	{A: "dDxOSK&L#6IbW5Y#EN1WFInnE", B: "ds*DnW2&~D_m~n/"},
	{A: "WlocCY1@qfDCcJ33*gF", B: "E7S@S^_yn6nRwjj~/4p3"},
	{A: "1TL4@ZKAK1dTVwa", B: "4XGvSLF_Tju^MOa"},
	{A: "QI$%5axEMo2xC@", B: "l4Yw~tx@dJxFH77k@0lr2Om&"},
	{A: "8yj5LKpkJ^tZrDhW/c", B: "0/lBB~jie*NR@4"},
	{A: "zT1-^#Qxnt", B: "5g_eaPF#Skpq8"},
	{A: "xpwkc$q4rw0TY1g5#E#", B: "97ATATaHu*sMih@o7Ws8cOQm"},
	{A: "0I0iJEM8RQk5Xv%M3uKj@TpU3gUE8", B: "%m*xGOZ94Eu"},
	{A: "LHc6&KUKY5JSUi$dvUR9m0i2&", B: "HkL5^UcFlJjPy^PoBjf&AS2fkRf"},
	{A: "biq~EH*1KjwM-ipT*ap_&Fv59S$#h", B: "HTi23^esLswn8#WmkeTn6"},
	{A: "U0K0SOenMQ$N&tAagsjsi-w", B: "ZBc%6d/C5^AYT"},
	{A: "vb9fwPYLF4S15", B: "7HNW9/M/bR"},
	{A: "INn8lj-97Gk&Q74Gk%nVnUBI3^fg", B: "Gp5B0^J2fD_A*n"},
	{A: "oPy6Z@6lW$rRS", B: "nc6a2DHMxYi"},
	{A: "rG8UGftTPkR", B: "@w9vhWjZ&*02%DXaJ^ApvMWX"},
	{A: "HnvTTWvHJ7h@G", B: "m8w/#DyOlcA8Gp"},
	{A: "3f$#I3$bT*3my", B: "pLDV#PaJt_saZ$RlG"},
	{A: "EP/EwEU@k5$*V/h2XOi-6E", B: "/VW%YP3GssrH&Hgf-hKDdFVQ"},
	{A: "^#7zFwQgEV3fXB8ArfWLi", B: "cHOgcl2Ui$/JdQoU_hETpHC"},
	{A: "VQvPKagE3V%xGfUjtX@yHaQ3U8bH", B: "fi9b^BLBu9tk2@&Ni"},
	{A: "8IDvYJ4vQM_q81", B: "G3WQ#vChjjNTaSC^PjLLP4^Pv1*z"},
	{A: "%hDhty7kCaL4t@N3h5/", B: "Hn69pWTJDrpnf@Ah"},
	{A: "EWIWuyeVb@OM*DmT$^B", B: "Zuy%FTI2jj$i@Uz0iD"},
	{A: "l/rYoL*Jp#g~a866i-~", B: "RPXogNVSTSl"},
	{A: "%TnHNM^qbKo4Ksrvyn6mit5^G", B: "5J$A$N86mSeQ1$F8f4gtg_ePi0"},
	{A: "TLBeuewVR*APRW", B: "Qne^ZJ~l$iliWNwazc~O0g"},
	{A: "ypVh_4gHcIjfK7EvQ@#46x3j", B: "PT27j1PJ@nz098sB3PUD"},
	{A: "TX-iLoZf-kwffOIRuXo", B: "8_s#zcd8PF"},
	{A: "PWAhjSuiSS9GtUp4Iu9~", B: "ix3xoqfIoct"},
	{A: "OZT5w*-sHEftYt2oalSUa3", B: "-zTddW%QfR891MN$R8URukNJAmI"},
	{A: "H$vafBY~%E6", B: "kzwLPxfF2m^Uxmbmy"},
	{A: "Xk2t-mMCRaC6V0GYbS-4eSTLxze@", B: "KAya$1Q5E*C/"},
	{A: "bogfWCiAp^JopF7kPD", B: "dCvLEjyr^_vbaEBA&Um~@NWKuu"},
	{A: "QXz@1j7n%5uS3t/~fSe01gQ#-iiQ", B: "zr/ppjs_F^Jk#W^z~S1"},
	{A: "E62TTgYO^2DzNmF5R_Otmq", B: "g3UcdHsB#yt-a9"},
	{A: "#rWEggwuAb/s96VUzT96fdk^7V", B: "fU2WYFYT_9x2pS8^Yp7PBjx-x"},
	{A: "uFWYkbG3ocIoPPWkudGj~piu5S#4", B: "7yL7O1DtmLQQmjr"},
	{A: "YuBQcG24y3VvIQ-1l9RW2xg@c", B: "ukMUJbn~kLo-/4S"},
	{A: "r$*lsgC5rer9R#7i", B: "GG8Fnv-Q#JuRb_rz98_XM#VdMaD&A"},
	{A: "L_0wV/5GumzmTVhbwZuBt2_6", B: "JOWWgkSCn-_NOcidEBaw"},
	{A: "cb0Fbn@oLYQN", B: "j2dA5o2R$NG9B"},
	{A: "LxVOQMTe14", B: "r%CJ2/uJH5zKLfO%#1PDbD*2*cym"},
	{A: "vn0d%L8Kxjrdl", B: "vAZ_WR&3g*#244ITR"},
	{A: "8hiEm/BU5iiQOO%U0CwX", B: "~81m_hbjW~SwuYG&dkwj1Z6wUURan"},
	{A: "o^K/3/rHGqr", B: "lvjpvHNNQ@LBN"},
	{A: "f2$e%UMl_JWh", B: "3eA/yTeK~M1yn4*xgt"},
	{A: "qFr8yXc@cNfl7Q15bUeI@v4zQeC27", B: "u3_POF*oQfOKZ1KsO~h"},
	{A: "GGj8*$PbC0_l-W^^", B: "2TOIF5ywA~cmcQ"},
	{A: "bS~lGE0J*gjAgUGc^bf81", B: "rhwXnW/$_Yd$8h$2~s7X&msj09"},
	{A: "RWuByUOFOFebP2DOIy7~znGQ", B: "Rctpv@riPPX5Crr"},
	{A: "esPCnEMZA3gGiG-i", B: "PxRYfOSZlT%k50N35"},
	{A: "Zp1MyR631ID5", B: "RpmG*3t@79ucYZd*OUGuzoymh7fQ"},
	{A: "savtmWezTn", B: "h1vs@K9m4QigH-RX"},
	{A: "xKiS-hyUwEW13perum", B: "MsGXolGhC$JR#0G"},
	{A: "km%hyF7JYL&@Cmjevl", B: "gHdRh_-ImPVpAGO*Yc*GJ-AMB2EI"},
	{A: "Wl5s9uOftwnf", B: "GV_WHrE4@I9IkwZwFzn^RO_0f"},
	{A: "ep~lzy6*9Z", B: "04M*O3mg/Qg/ryRj2V7aw"},
	{A: "GAuul^sQAP#V7A", B: "GoC3g@nwISXQiFov*_Io"},
	{A: "VWWu/SRk9ko1#&l#-q3uIo", B: "/dbP/j~OWn0"},
	{A: "_9D#/w-tcvH#@SgRghxK2W2Z#", B: "Xa2D~wwhSay1R0m%3$@"},
	{A: "B@7ms%n3Un%Mxse6MqK28pu~", B: "5W3BJPKiqZ"},
	{A: "Cszx*ti1rUJ", B: "1nZJ4aWAsX%jO62i"},
	{A: "u1vIqg8F9*2%q9RcQJL4%eP", B: "1CzmO~yfX5@wF^c6re"},
	{A: "p*aKNC4rldPfkb_fcpul#", B: "uGaKk4cQRQc3L1@1p4tXsQhSQ"},
	{A: "CXL8N@xkBgJxwF007thG%W~FHqpTl", B: "vRN0SXr*8XHEA4@HiJPl7Yxzaq0Zf"},
	{A: "Yg7/l#O-#HqOqLwer", B: "gYofEGFot7nvX"},
	{A: "Iid2TQ2~FMxCF#wNb~", B: "ObHRiL2a1J2fszFC~onQ"},
	{A: "qd^vfES^-W", B: "FvBlSdIVvAx%6a@ylWLXI%"},
	{A: "9J78_6C2yu65^I2Z/v3K$*v", B: "Pdipy$Eg@EmkPgBqE@P&z"},
	{A: "NC0fY_cw/w", B: "twLfY_cm9v"},
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
	{A: "lbwWi@35Cpj5xDRPTKo&UDTPd%irr", B: "Qu-qIyt#Eb%V1AHkq4d5dCHHj4", Skip: true},
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
			t.Logf("same = %v %v", len(same), same)
			t.Logf("diff = %v %v", len(diff), diff)
			t.Logf("len1=%v\ta1=%v\tb1=%v\tin1=%q", len(v.A), state1.a, state1.b, v.A)
			t.Logf("len2=%v\ta2=%v\tb2=%v\tin2=%q", len(v.B), state2.a, state2.b, v.B)

			var h1, h2 [4]uint64
			var a1, a2 [4]uint64
			for i := uint64(0); i < 256; i += 32 {
				a1[0], a2[0] = state1.Uint64LE(state1.a+i+0), state2.Uint64LE(state2.a+i+0)
				a1[1], a2[1] = state1.Uint64LE(state1.a+i+8), state2.Uint64LE(state2.a+i+8)
				a1[2], a2[2] = state1.Uint64LE(state1.a+i+16), state2.Uint64LE(state2.a+i+16)
				a1[3], a2[3] = state1.Uint64LE(state1.a+i+24), state2.Uint64LE(state2.a+i+24)

				h1[0], h2[0] = h1[3]+a1[0], h2[3]+a2[0]
				h1[1], h2[1] = h1[0]^a1[1], h2[0]^a2[1]
				h1[2], h2[2] = h1[1]+a1[2], h2[1]+a2[2]
				h1[3], h2[3] = h1[2]*a1[3], h2[2]*a2[3]

				t.Logf("a1=%v\ta2=%v\th1=%v\th2=%v\t%v", a1[0], a2[0], h1[0], h2[0], h1[0] == h2[0])
				t.Logf("a1=%v\ta2=%v\th1=%v\th2=%v\t%v", a1[1], a2[1], h1[1], h2[1], h1[1] == h2[1])
				t.Logf("a1=%v\ta2=%v\th1=%v\th2=%v\t%v", a1[2], a2[2], h1[2], h2[2], h1[2] == h2[2])
				t.Logf("a1=%v\ta2=%v\th1=%v\th2=%v\t%v", a1[3], a2[3], h1[3], h2[3], h1[3] == h2[3])

				t.Logf("###")
			}
			assert.Assert(t, v.Skip, res1)
		}
	}
}
