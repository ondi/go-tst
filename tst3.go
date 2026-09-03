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
	state   [256]U64_iv_t
	a, b, e uint64
}

func (self *State256_t) Reset() {
	self.state = NewInv64()
	self.a = 0
	self.e = 0b10000000_01000000_00100000_00010000_00001000_00000100_00000010_00000001
}

func (self *State256_t) StateAdd1(c byte) uint64 {
	self.a = (self.a + 1) % 256
	self.b = (self.a + 2*(self.b+self.state[c].A) + 1) % 256
	self.e = RORL64N((self.e^self.state[self.b].B), self.state[self.b].D) * self.state[self.b].C
	self.state[self.b].B += 512
	self.state[self.b].C = InvUint64(self.state[self.b].B)
	// self.state[self.b].D = self.state[self.b].D%63 + 1
	self.state[self.b].D = self.state[self.b].D%49 + 8
	self.state[self.a], self.state[self.b] = self.state[self.b], self.state[self.a]
	return self.e
}

func (self *State256_t) StateAdd2(c byte) uint64 {
	self.a = (self.a + 1) % 256
	self.b = (self.a + 2*(self.b+self.state[c].A) + 1) % 256
	self.e = ROL64N((self.e^self.a), 4) * (self.a + self.b)
	self.e = ROL64N((self.e^self.b), 4) * (self.a ^ self.b)
	self.state[self.a], self.state[self.b] = self.state[self.b], self.state[self.a]
	return self.e
}

func (self *State256_t) StateAdd(c byte) uint64 {
	self.a = (self.a + 1) % 256
	self.b = (self.a + 2*(self.b+self.state[c].A) + 1) % 256
	self.e = ROL64N(self.e^(self.a+self.b), 7) * (self.a ^ self.b)
	self.state[self.a], self.state[self.b] = self.state[self.b], self.state[self.a]
	return self.e
}

func (self *State256_t) StateAdd4(c byte) uint64 {
	self.a = (self.a + 1) % 256
	self.b = (self.a + 2*(self.b+self.state[c].A) + 1) % 256
	a := ROL64N((self.e^self.a), 9) * (self.a ^ self.b)
	b := ROR64N((self.e^self.b), 9) * (self.a + self.b)
	self.e = a ^ b
	self.state[self.a], self.state[self.b] = self.state[self.b], self.state[self.a]
	return self.e
}

func (self *State256_t) Sum64() uint64 {
	return self.e
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

func RORL64N(in, n uint64) (out uint64) {
	if n&1 > 0 {
		return (in >> n) | (in << (64 - n))
	} else if n > 0 {
		return (in << n) | (in >> (64 - n))
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
