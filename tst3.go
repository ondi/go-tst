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
	var sx, sy, vx, vy uint64
	key := key3_t{}
	state := State256_t{}
	state.Reset()
	for i, key.code = range []byte(prefix) {
		key.pos = int32(i)
		sx, sy, vx, vy = state.StateNext(key.code)
		key.hash = (key.hash*sx + vx + 1) * (key.hash*sy + vy + 1)
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
	var sx, sy, vx, vy uint64
	var temp *Mapped_t[Value_t]
	key := key3_t{}
	state := State256_t{}
	state.Reset()
	for length, key.code = range []byte(in) {
		key.pos = int32(length)
		sx, sy, vx, vy = state.StateNext(key.code)
		key.hash = (key.hash*sx + vx + 1) * (key.hash*sy + vy + 1)
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
	state [256]uint8
	x, y  uint64
}

func (self *State256_t) Reset() {
	self.state = [256]uint8{
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
}

func (self *State256_t) StateNext(in byte) (uint64, uint64, uint64, uint64) {
	self.x = (self.x + 1) % 256
	self.y = (self.y + uint64(self.state[self.x]) + uint64(in) + 1) % 256
	self.state[self.x], self.state[self.y] = self.state[self.y], self.state[self.x]
	return uint64(self.state[self.x]), uint64(self.state[self.y]), self.x, self.y
}

type StateHash_t struct {
	state State256_t
	hash  uint64
}

func NewStateHash() (self *StateHash_t) {
	self = &StateHash_t{}
	self.state.Reset()
	return
}

func (self *StateHash_t) Reset() {
	self.state.Reset()
	self.hash = 0
}

func (self *StateHash_t) Sum64(in []byte) uint64 {
	var sx, sy, vx, vy uint64
	for _, code := range in {
		sx, sy, vx, vy = self.state.StateNext(code)
		// self.hash = (self.hash*sx + vx) * (self.hash*sy + vy)
		self.hash = (self.hash*sx + vx + 1) * (self.hash*sy + vy + 1)
	}
	return self.hash
}
