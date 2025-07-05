//
// ternary search tree
//

package tst

type key3_t struct {
	hash uint64
	pos  int32
	code byte
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
	var i int
	key := key3_t{}
	state := State256_t{}
	state.Reset()
	for i, key.code = range []byte(prefix) {
		key.pos = int32(i)
		state.StateNext(key.code)
		key.hash = state.Sum64()
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
	var temp *Mapped_t[Value_t]
	key := key3_t{}
	state := State256_t{}
	state.Reset()
	for length, key.code = range []byte(in) {
		key.pos = int32(length)
		state.StateNext(key.code)
		key.hash = state.Sum64()
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
	state   [256]uint64
	x, y, z uint64
}

func (self *State256_t) Reset() {
	self.state = [256]uint64{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
		32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47,
		48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63,
		64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79,
		80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95,
		96, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111,
		112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127,
		128, 129, 130, 131, 132, 133, 134, 135, 136, 137, 138, 139, 140, 141, 142, 143,
		144, 145, 146, 147, 148, 149, 150, 151, 152, 153, 154, 155, 156, 157, 158, 159,
		160, 161, 162, 163, 164, 165, 166, 167, 168, 169, 170, 171, 172, 173, 174, 175,
		176, 177, 178, 179, 180, 181, 182, 183, 184, 185, 186, 187, 188, 189, 190, 191,
		192, 193, 194, 195, 196, 197, 198, 199, 200, 201, 202, 203, 204, 205, 206, 207,
		208, 209, 210, 211, 212, 213, 214, 215, 216, 217, 218, 219, 220, 221, 222, 223,
		224, 225, 226, 227, 228, 229, 230, 231, 232, 233, 234, 235, 236, 237, 238, 239,
		240, 241, 242, 243, 244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255,
	}
	self.x = 0
	self.y = 0
	self.z = 0
}

func (self *State256_t) StateNext(in byte) {
	self.x = (self.x + 1) % 256
	self.y = (self.state[self.y] + self.state[self.x] + self.state[in] + 1) % 256
	self.z = (self.state[self.z] + self.state[self.y] + uint64(in) + 1) % 256
	self.state[self.x], self.state[self.y], self.state[self.z] = self.state[self.y], self.state[self.z], self.state[self.x]
}

func (self *State256_t) Sum64() (res uint64) {
	x := Backward(256, self.x, 8)
	y := Backward(256, self.y, 8)
	z := Backward(256, self.z, 8)

	res = (self.state[(z+1)%256]<<(8*0) |
		self.state[(z+2)%256]<<(8*1) |
		self.state[(z+3)%256]<<(8*2) |
		self.state[(z+4)%256]<<(8*3) |
		self.state[(z+5)%256]<<(8*4) |
		self.state[(z+6)%256]<<(8*5) |
		self.state[(z+7)%256]<<(8*6) |
		self.state[(z+8)%256]<<(8*7))

	res = res + (self.state[(y+1)%256]<<(8*0) |
		self.state[(y+2)%256]<<(8*1) |
		self.state[(y+3)%256]<<(8*2) |
		self.state[(y+4)%256]<<(8*3) |
		self.state[(y+5)%256]<<(8*4) |
		self.state[(y+6)%256]<<(8*5) |
		self.state[(y+7)%256]<<(8*6) |
		self.state[(y+8)%256]<<(8*7))

	res = res * (self.state[(x+1)%256]<<(8*0) |
		self.state[(x+2)%256]<<(8*1) |
		self.state[(x+3)%256]<<(8*2) |
		self.state[(x+4)%256]<<(8*3) |
		self.state[(x+5)%256]<<(8*4) |
		self.state[(x+6)%256]<<(8*5) |
		self.state[(x+7)%256]<<(8*6) |
		self.state[(x+8)%256]<<(8*7))

	return
}

func Forward(size uint64, current uint64, offset uint64) uint64 {
	return (size + current + offset) % size
}

func Backward(size uint64, current uint64, offset uint64) uint64 {
	return (size + current - offset) % size
}

func StateSum64(in []byte) uint64 {
	var state State256_t
	state.Reset()
	for _, code := range in {
		state.StateNext(code)
	}
	return state.Sum64()
}
