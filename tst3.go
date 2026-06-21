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

type State256_t struct {
	state   [256]*U64_iv_t
	a, b, e uint64
}

func (self *State256_t) Reset() {
	self.state = iv64_1
	self.a = 255
	self.b = 0
	self.e = 0b10000000_01000000_00100000_00010000_00001000_00000100_00000010_00000000
}

// 1000: 9, 188h
func (self *State256_t) StateAdd01(in byte) uint64 {
	self.a = (self.a + 1) % 256
	self.b = (self.b + 2*(self.a+self.state[in].A) + 1) % 256
	// self.a + self.b always odd
	self.e = ROL64N((self.e^self.state[self.a].A)*self.state[self.b].B, 4)
	self.e = ROR64N((self.e^self.state[self.b].A), 4) * self.state[self.b].C
	self.state[self.a], self.state[self.b] = self.state[self.b], self.state[self.a]
	return self.e
}

// 1000: 9, 160h
func (self *State256_t) StateAdd02(in byte) uint64 {
	self.a = (self.a + 1) % 256
	self.b = (self.b + 2*(self.a+self.state[in].A) + 1) % 256
	// self.a + self.b always odd
	self.e = ROR64N((self.e^self.state[self.b].A)*self.state[self.b].B, 1) ^ self.state[self.b].C
	self.state[self.a], self.state[self.b] = self.state[self.b], self.state[self.a]
	return self.e
}

// odd*odd=odd, even*odd=even, even*even=even
func (self *State256_t) StateAdd(in byte) uint64 {
	self.a = (self.a + 1) % 256
	self.b = (self.b + 2*(self.a+self.state[in].A) + 1) % 256
	// self.a + self.b always odd
	self.e = ROR64N((self.e^self.state[self.b].B), 1) * self.state[self.b].C
	self.state[self.a], self.state[self.b] = self.state[self.b], self.state[self.a]
	return self.e
}

func (self *State256_t) Sum64() uint64 {
	return self.e
}

// n = [0,63]
func ROL64N(in uint64, n uint64) (out uint64) {
	if n > 0 {
		return (in << n) | (in >> (64 - n))
	}
	return in
}

// n = [0,63]
func ROR64N(in uint64, n uint64) (out uint64) {
	if n > 0 {
		return (in >> n) | (in << (64 - n))
	}
	return in
}

func ROL64B(in uint64, mask uint64, bits uint64) (out uint64) {
	if in > 0 {
		for in&mask != bits {
			in = (in << 1) | (in >> 63)
		}
	}
	return in
}

func ROR64B(in uint64, mask uint64, bits uint64) (out uint64) {
	if in > 0 {
		for in&mask != bits {
			in = (in >> 1) | (in << 63)
		}
	}
	return in
}

func Min(mod uint64, min uint64, a uint64, in ...uint64) uint64 {
	a = a % mod
	for _, b := range in {
		if b = b % mod; b < a && b > min {
			a = b
		}
	}
	if a > min {
		return a
	}
	return min
}

func Max(mod uint64, min uint64, a uint64, in ...uint64) uint64 {
	a = a % mod
	for _, b := range in {
		if b = b % mod; b > a && b > min {
			a = b
		}
	}
	if a > min {
		return a
	}
	return min
}

func Forward(size uint64, current uint64, offset uint64) uint64 {
	return (size + current + offset%size) % size
}

func Backward(size uint64, current uint64, offset uint64) uint64 {
	return (size + current - offset%size) % size
}

func Mul_u64(a uint64, b uint64) (hi uint64, lo uint64) {
	a_hi, b_hi := a>>32, b>>32
	a_lo, b_lo := a&0xFFFFFFFF, b&0xFFFFFFFF

	a_hi_b_hi := a_hi * b_hi
	a_hi_b_lo := a_hi * b_lo
	a_lo_b_hi := a_lo * b_hi
	a_lo_b_lo := a_lo * b_lo

	intermediate := a_hi_b_lo + a_lo_b_hi + (a_lo_b_lo >> 32)

	hi = a_hi_b_hi + (intermediate >> 32)
	lo = (intermediate << 32) | (a_lo_b_lo & 0xFFFFFFFF)

	return
}

// 0x12345678 <-> []{0x78, 0x56, 0x34, 0x12}
func (self *State256_t) Uint64LE(i uint64, step uint64) uint64 {
	return self.state[(i+0*step)%256].A<<(8*0) |
		self.state[(i+1*step)%256].A<<(8*1) |
		self.state[(i+2*step)%256].A<<(8*2) |
		self.state[(i+3*step)%256].A<<(8*3) |
		self.state[(i+4*step)%256].A<<(8*4) |
		self.state[(i+5*step)%256].A<<(8*5) |
		self.state[(i+6*step)%256].A<<(8*6) |
		self.state[(i+7*step)%256].A<<(8*7)
}

// 0x12345678 <-> []{0x12, 0x34, 0x56, 0x78}
func (self *State256_t) Uint64BE(i uint64, step uint64) uint64 {
	return self.state[(i+0*step)%256].A<<(8*7) |
		self.state[(i+1*step)%256].A<<(8*6) |
		self.state[(i+2*step)%256].A<<(8*5) |
		self.state[(i+3*step)%256].A<<(8*4) |
		self.state[(i+4*step)%256].A<<(8*3) |
		self.state[(i+5*step)%256].A<<(8*2) |
		self.state[(i+6*step)%256].A<<(8*1) |
		self.state[(i+7*step)%256].A<<(8*0)
}
