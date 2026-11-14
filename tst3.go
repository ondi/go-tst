//
// ternary search tree
//

package tst

type key3_t struct {
	prev uint64
	next uint64
	pos  int
}

type Mapped_t[Value_t any] struct {
	value Value_t
}

type Tree3_t[Value_t any] struct {
	root map[key3_t]*Mapped_t[Value_t]
}

func NewTree3[Value_t any]() *Tree3_t[Value_t] {
	return &Tree3_t[Value_t]{
		root: map[key3_t]*Mapped_t[Value_t]{},
	}
}

func (self *Tree3_t[Value_t]) Add(prefix string, value Value_t) (mapped *Mapped_t[Value_t], ok bool) {
	var code byte
	key := key3_t{}
	state := State256_t{}
	state.Reset()
	for key.pos, code = range []byte(prefix) {
		key.prev = state.Sum64()
		key.next = state.StateAdd(code)
		if mapped, ok = self.root[key]; !ok {
			self.root[key] = nil
		}
	}
	if mapped == nil {
		mapped = &Mapped_t[Value_t]{value: value}
		self.root[key] = mapped
		ok = true
	}
	return
}

func (self *Tree3_t[Value_t]) Search(in string) (value Value_t, length int, found int) {
	var ok bool
	var code byte
	var temp *Mapped_t[Value_t]
	key := key3_t{}
	state := State256_t{}
	state.Reset()
	for length, code = range []byte(in) {
		key.pos = length
		key.prev = state.Sum64()
		key.next = state.StateAdd(code)
		if temp, ok = self.root[key]; !ok {
			return
		}
		if temp != nil {
			found++
			value = temp.value
		}
	}
	return
}

type Cell_t struct {
	A, B, C, D uint64
	E          uint64
}

type State256_t struct {
	state   [256]Cell_t
	a, b, e uint64
}

func (self *State256_t) Reset() {
	// self.state = [256]Cell_t{
	// 	{A: 0}, {A: 1}, {A: 2}, {A: 3}, {A: 4}, {A: 5}, {A: 6}, {A: 7}, {A: 8}, {A: 9}, {A: 10}, {A: 11}, {A: 12}, {A: 13}, {A: 14}, {A: 15}, {A: 16}, {A: 17}, {A: 18}, {A: 19}, {A: 20}, {A: 21}, {A: 22}, {A: 23}, {A: 24}, {A: 25}, {A: 26}, {A: 27}, {A: 28}, {A: 29}, {A: 30}, {A: 31}, {A: 32}, {A: 33}, {A: 34}, {A: 35}, {A: 36}, {A: 37}, {A: 38}, {A: 39}, {A: 40}, {A: 41}, {A: 42}, {A: 43}, {A: 44}, {A: 45}, {A: 46}, {A: 47}, {A: 48}, {A: 49}, {A: 50}, {A: 51}, {A: 52}, {A: 53}, {A: 54}, {A: 55}, {A: 56}, {A: 57}, {A: 58}, {A: 59}, {A: 60}, {A: 61}, {A: 62}, {A: 63}, {A: 64}, {A: 65}, {A: 66}, {A: 67}, {A: 68}, {A: 69}, {A: 70}, {A: 71}, {A: 72}, {A: 73}, {A: 74}, {A: 75}, {A: 76}, {A: 77}, {A: 78}, {A: 79}, {A: 80}, {A: 81}, {A: 82}, {A: 83}, {A: 84}, {A: 85}, {A: 86}, {A: 87}, {A: 88}, {A: 89}, {A: 90}, {A: 91}, {A: 92}, {A: 93}, {A: 94}, {A: 95}, {A: 96}, {A: 97}, {A: 98}, {A: 99}, {A: 100}, {A: 101}, {A: 102}, {A: 103}, {A: 104}, {A: 105}, {A: 106}, {A: 107}, {A: 108}, {A: 109}, {A: 110}, {A: 111}, {A: 112}, {A: 113}, {A: 114}, {A: 115}, {A: 116}, {A: 117}, {A: 118}, {A: 119}, {A: 120}, {A: 121}, {A: 122}, {A: 123}, {A: 124}, {A: 125}, {A: 126}, {A: 127}, {A: 128}, {A: 129}, {A: 130}, {A: 131}, {A: 132}, {A: 133}, {A: 134}, {A: 135}, {A: 136}, {A: 137}, {A: 138}, {A: 139}, {A: 140}, {A: 141}, {A: 142}, {A: 143}, {A: 144}, {A: 145}, {A: 146}, {A: 147}, {A: 148}, {A: 149}, {A: 150}, {A: 151}, {A: 152}, {A: 153}, {A: 154}, {A: 155}, {A: 156}, {A: 157}, {A: 158}, {A: 159}, {A: 160}, {A: 161}, {A: 162}, {A: 163}, {A: 164}, {A: 165}, {A: 166}, {A: 167}, {A: 168}, {A: 169}, {A: 170}, {A: 171}, {A: 172}, {A: 173}, {A: 174}, {A: 175}, {A: 176}, {A: 177}, {A: 178}, {A: 179}, {A: 180}, {A: 181}, {A: 182}, {A: 183}, {A: 184}, {A: 185}, {A: 186}, {A: 187}, {A: 188}, {A: 189}, {A: 190}, {A: 191}, {A: 192}, {A: 193}, {A: 194}, {A: 195}, {A: 196}, {A: 197}, {A: 198}, {A: 199}, {A: 200}, {A: 201}, {A: 202}, {A: 203}, {A: 204}, {A: 205}, {A: 206}, {A: 207}, {A: 208}, {A: 209}, {A: 210}, {A: 211}, {A: 212}, {A: 213}, {A: 214}, {A: 215}, {A: 216}, {A: 217}, {A: 218}, {A: 219}, {A: 220}, {A: 221}, {A: 222}, {A: 223}, {A: 224}, {A: 225}, {A: 226}, {A: 227}, {A: 228}, {A: 229}, {A: 230}, {A: 231}, {A: 232}, {A: 233}, {A: 234}, {A: 235}, {A: 236}, {A: 237}, {A: 238}, {A: 239}, {A: 240}, {A: 241}, {A: 242}, {A: 243}, {A: 244}, {A: 245}, {A: 246}, {A: 247}, {A: 248}, {A: 249}, {A: 250}, {A: 251}, {A: 252}, {A: 253}, {A: 254}, {A: 255},
	// }
	self.e = 0b10010000_01001000_00100100_00010010_10001000_01000100_00100010_00010001
}

func (self *State256_t) StateAdd(c byte) uint64 {
	self.a = (self.a + 1) & 0xFF
	self.b = (self.b + self.a + self.e + uint64(c) + 1) & 0xFF
	v := Shifts[self.a%8]
	B, C := CellBC(v.Bits, self.b)
	self.e = ROL64N(ROL64N(self.e^B, v.Bits)*C, v.Bits)
	// self.state[self.b].E++
	// self.state[self.a], self.state[self.b] = self.state[self.b], self.state[self.a]
	return self.e
}

func (self *State256_t) Sum64() uint64 {
	return self.e
}

type Shift_t struct {
	Bits uint64
}

var Shifts = []Shift_t{
	{Bits: 17},
	{Bits: 16},
	{Bits: 15},
	{Bits: 14},
	{Bits: 13},
	{Bits: 12},
	{Bits: 11},
	{Bits: 10},
}

func CellBC(bits uint64, pos uint64) (b uint64, c uint64) {
	var min uint64 = 1 << (bits - 1)
	var max uint64 = (1 << bits) - 1
	min++
	step := (max - min) / (256 - 1)
	k := min + step*pos
	b = (1<<(64-bits))*k + 1
	c = (max+1-k)<<(64-bits) + 1
	return
}

// a * InvUint64(a) = 1
// только для нечётных чисел
// обратный элемент к a по модулю 2^64 методом Ньютона
func InvUint64(a uint64) (x uint64) {
	x = 1
	for range 6 {
		x *= 2 - a*x
	}
	return
}

// n = [0,63]
func ROL64N(in, n uint64) uint64 {
	if n > 0 {
		return (in << n) | (in >> (64 - n))
	}
	return in
}

// n = [0,63]
func ROR64N(in, n uint64) uint64 {
	if n > 0 {
		return (in >> n) | (in << (64 - n))
	}
	return in
}

func Forward(size uint64, current uint64, offset uint64) uint64 {
	return (size + current + offset%size) % size
}

func Backward(size uint64, current uint64, offset uint64) uint64 {
	return (size + current - offset%size) % size
}

func Mul_u64(a uint64, b uint64) (hi uint64, lo uint64) {
	a_hi, a_lo := a>>32, a&0xFFFFFFFF
	b_hi, b_lo := b>>32, b&0xFFFFFFFF

	lo = a_lo * b_lo
	mid := a_hi*b_lo + a_lo*b_hi + (lo >> 32)

	hi = a_hi*b_hi + (mid >> 32)
	lo = (mid << 32) | (lo & 0xFFFFFFFF)

	return
}

// Little Endian: 0x12345678 <-> []{0x78, 0x56, 0x34, 0x12}
// Big Endian:    0x12345678 <-> []{0x12, 0x34, 0x56, 0x78}
