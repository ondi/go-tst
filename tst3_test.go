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
		hx := StateSum64(buf)
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
	{"LHc6&KUKY5JSUi$dvUR9m0i2&", "HkL5^UcFlJjPy^PoBjf&AS2fkRf"},
	{"biq~EH*1KjwM-ipT*ap_&Fv59S$#h", "HTi23^esLswn8#WmkeTn6"},
	{"U0K0SOenMQ$N&tAagsjsi-w", "ZBc%6d/C5^AYT"},
	{"vb9fwPYLF4S15", "7HNW9/M/bR"},
	{"INn8lj-97Gk&Q74Gk%nVnUBI3^fg", "Gp5B0^J2fD_A*n"},
	{"oPy6Z@6lW$rRS", "nc6a2DHMxYi"},
	{"rG8UGftTPkR", "@w9vhWjZ&*02%DXaJ^ApvMWX"},
	{"HnvTTWvHJ7h@G", "m8w/#DyOlcA8Gp"},
	{"3f$#I3$bT*3my", "pLDV#PaJt_saZ$RlG"},
	{"EP/EwEU@k5$*V/h2XOi-6E", "/VW%YP3GssrH&Hgf-hKDdFVQ"},
	{"^#7zFwQgEV3fXB8ArfWLi", "cHOgcl2Ui$/JdQoU_hETpHC"},
	{"VQvPKagE3V%xGfUjtX@yHaQ3U8bH", "fi9b^BLBu9tk2@&Ni"},
	{"8IDvYJ4vQM_q81", "G3WQ#vChjjNTaSC^PjLLP4^Pv1*z"},
	{"%hDhty7kCaL4t@N3h5/", "Hn69pWTJDrpnf@Ah"},
	{"EWIWuyeVb@OM*DmT$^B", "Zuy%FTI2jj$i@Uz0iD"},
	{"l/rYoL*Jp#g~a866i-~", "RPXogNVSTSl"},
	{"%TnHNM^qbKo4Ksrvyn6mit5^G", "5J$A$N86mSeQ1$F8f4gtg_ePi0"},
	{"TLBeuewVR*APRW", "Qne^ZJ~l$iliWNwazc~O0g"},
	{"ypVh_4gHcIjfK7EvQ@#46x3j", "PT27j1PJ@nz098sB3PUD"},
	{"TX-iLoZf-kwffOIRuXo", "8_s#zcd8PF"},
	{"PWAhjSuiSS9GtUp4Iu9~", "ix3xoqfIoct"},
	{"OZT5w*-sHEftYt2oalSUa3", "-zTddW%QfR891MN$R8URukNJAmI"},
	{"H$vafBY~%E6", "kzwLPxfF2m^Uxmbmy"},
	{"Xk2t-mMCRaC6V0GYbS-4eSTLxze@", "KAya$1Q5E*C/"},
	{"bogfWCiAp^JopF7kPD", "dCvLEjyr^_vbaEBA&Um~@NWKuu"},
	{"QXz@1j7n%5uS3t/~fSe01gQ#-iiQ", "zr/ppjs_F^Jk#W^z~S1"},
	{"E62TTgYO^2DzNmF5R_Otmq", "g3UcdHsB#yt-a9"},
	{"#rWEggwuAb/s96VUzT96fdk^7V", "fU2WYFYT_9x2pS8^Yp7PBjx-x"},
	{"uFWYkbG3ocIoPPWkudGj~piu5S#4", "7yL7O1DtmLQQmjr"},
	{"YuBQcG24y3VvIQ-1l9RW2xg@c", "ukMUJbn~kLo-/4S"},
	{"r$*lsgC5rer9R#7i", "GG8Fnv-Q#JuRb_rz98_XM#VdMaD&A"},
	{"L_0wV/5GumzmTVhbwZuBt2_6", "JOWWgkSCn-_NOcidEBaw"},
	{"cb0Fbn@oLYQN", "j2dA5o2R$NG9B"},
	{"LxVOQMTe14", "r%CJ2/uJH5zKLfO%#1PDbD*2*cym"},
	{"vn0d%L8Kxjrdl", "vAZ_WR&3g*#244ITR"},
	{"8hiEm/BU5iiQOO%U0CwX", "~81m_hbjW~SwuYG&dkwj1Z6wUURan"},
	{"o^K/3/rHGqr", "lvjpvHNNQ@LBN"},
	{"f2$e%UMl_JWh", "3eA/yTeK~M1yn4*xgt"},
	{"qFr8yXc@cNfl7Q15bUeI@v4zQeC27", "u3_POF*oQfOKZ1KsO~h"},
	{"GGj8*$PbC0_l-W^^", "2TOIF5ywA~cmcQ"},
	{"bS~lGE0J*gjAgUGc^bf81", "rhwXnW/$_Yd$8h$2~s7X&msj09"},
	{"RWuByUOFOFebP2DOIy7~znGQ", "Rctpv@riPPX5Crr"},
	{"esPCnEMZA3gGiG-i", "PxRYfOSZlT%k50N35"},
	{"Zp1MyR631ID5", "RpmG*3t@79ucYZd*OUGuzoymh7fQ"},
	{"savtmWezTn", "h1vs@K9m4QigH-RX"},
	{"xKiS-hyUwEW13perum", "MsGXolGhC$JR#0G"},
	{"km%hyF7JYL&@Cmjevl", "gHdRh_-ImPVpAGO*Yc*GJ-AMB2EI"},
	{"Wl5s9uOftwnf", "GV_WHrE4@I9IkwZwFzn^RO_0f"},
	{"ep~lzy6*9Z", "04M*O3mg/Qg/ryRj2V7aw"},
	{"GAuul^sQAP#V7A", "GoC3g@nwISXQiFov*_Io"},
	{"VWWu/SRk9ko1#&l#-q3uIo", "/dbP/j~OWn0"},
	{"_9D#/w-tcvH#@SgRghxK2W2Z#", "Xa2D~wwhSay1R0m%3$@"},
	{"B@7ms%n3Un%Mxse6MqK28pu~", "5W3BJPKiqZ"},
	{"Cszx*ti1rUJ", "1nZJ4aWAsX%jO62i"},
	{"u1vIqg8F9*2%q9RcQJL4%eP", "1CzmO~yfX5@wF^c6re"},
	{"p*aKNC4rldPfkb_fcpul#", "uGaKk4cQRQc3L1@1p4tXsQhSQ"},
	{"CXL8N@xkBgJxwF007thG%W~FHqpTl", "vRN0SXr*8XHEA4@HiJPl7Yxzaq0Zf"},
	{"Yg7/l#O-#HqOqLwer", "gYofEGFot7nvX"},
	{"Iid2TQ2~FMxCF#wNb~", "ObHRiL2a1J2fszFC~onQ"},
	{"qd^vfES^-W", "FvBlSdIVvAx%6a@ylWLXI%"},
	{"9J78_6C2yu65^I2Z/v3K$*v", "Pdipy$Eg@EmkPgBqE@P&z"},
	{"NC0fY_cw/w", "twLfY_cm9v"},
	{"6%qZjW33_o72CQwBsltF", "5KR5WmyG/J&z_M8hH"},
	{"lLD5B-DMn0oaljG78Aqg^b", "hIrdHtXM#87X_$iKMJ"},
	{"*gCBMAxWCIP%2YnKO*9I4W", "5YY$I@VK$5Djd-wEslLyQW/4"},
	{"Ig1HiBNVf2MO/F", "kL/@uK3sZ3_&GKebCT"},
	{"sw@&ZK^#by", "Jzf&ZK^#by"},
	{"jiFA5&nMvpOP1K", "1OT#YhDMvpOP1K"},
	{"iikgD2qmG/", "3B4eD2qmG/"},
	{"Ueh_M8qKvhnm0Y49/9B%Aj&VwI", "j#_B@MWopK"},
	{"RmPsw*HWdCOvK", "XwPnomWl^ZRScMo&6~GP9"},
	{"O0T_f92xgZPd/O_@", "Lu-&P71hPsHyqf$f"},
	{"%dBBzCq3WO", "eA@prdCZz20mUr@OvAE1"},
	{"cXUmas9lVc0Xz_IXabCMr3NymV", "FA%$o1D$ELYBaaGDBAcDVIC"},
	{"t&oXeZB*qz0#%%0t", "#_K#-hu$/s2V/-~t"},
	{"UEj3&G71/6CzCiorg", "@H/WHp9y2jiQensCwGWT_EX_il"},
	{"gG&2qwirJ6GGDLiH-bWY%7pRTAdpS", "4zDVz%c%$Z@8"},
	{"Dr5_r2^Ti0S1Z", "f2xWnyf4IFksld%j4svW5phJK"},
	{"6^/0uQ*rdiQg5k3y8lB&m", "I4g3LM4yfE@oEHeq/lE"},
	{"V0aUwgUlvo", "yQw1yJ3PMMb&o#Y"},
	{"pGyZaWDx#_S", "&JT&bt%IJsUjV@$i6@3M"},
	{"0DyZAIUSmvE-kWkrb13F$", "i^VBiJD_/HaXERkh*3BVD225^Gd2"},
	{"kq8OZPR3byKOjHRl42C*pJtq1u", "eIoNZUKWfy^z$D13V2M*5"},
}

func Test_Tst3_04(t *testing.T) {
	for _, v := range in {
		res1 := StateSum64([]byte(v[0]))
		res2 := StateSum64([]byte(v[1]))
		assert.Assert(t, res1 != res2, fmt.Sprintf("%v %q %q", res1, v[0], v[1]))
	}
}
