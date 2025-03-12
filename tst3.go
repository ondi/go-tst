//
// ternary search tree
//

package tst

import "encoding/binary"

const (
	FnvOffset64 = 14695981039346656037
	FnvPrime64  = 1099511628211
)

type key3_t struct {
	hash uint64
	pos  int32
	code int32
}

type mapped3_t[Value_t any] struct {
	value Value_t
}

type Tree3_t[Value_t any] struct {
	root map[key3_t]*mapped3_t[Value_t]
}

func NewTree3[Value_t any]() *Tree3_t[Value_t] {
	return &Tree3_t[Value_t]{
		root: map[key3_t]*mapped3_t[Value_t]{},
	}
}

func (self *Tree3_t[Value_t]) Add(prefix string, value Value_t) (ok bool) {
	var i int
	state := NewState8()
	key := key3_t{hash: FnvOffset64}
	var temp *mapped3_t[Value_t]
	for i, key.code = range prefix {
		key.pos = int32(i)
		key.hash ^= uint64(key.code)
		key.hash *= FnvPrime64
		key.hash = ReplaceUint64(state, key.hash)
		if temp, ok = self.root[key]; !ok {
			self.root[key] = nil
		}
	}
	if temp == nil {
		self.root[key] = &mapped3_t[Value_t]{value: value}
		return true
	}
	return false
}

func (self *Tree3_t[Value_t]) Search(in string) (value Value_t, length int, found int) {
	var ok bool
	var temp *mapped3_t[Value_t]
	state := NewState8()
	key := key3_t{hash: FnvOffset64}
	for length, key.code = range in {
		key.pos = int32(length)
		key.hash ^= uint64(key.code)
		key.hash *= FnvPrime64
		key.hash = ReplaceUint64(state, key.hash)
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

type State8_t struct {
	state [256]uint8
	x, y  int
}

func NewState8() (self *State8_t) {
	self = &State8_t{}
	self.Reset()
	return
}

func (self *State8_t) Reset() {
	self.state = [256]uint8{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
		10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
		20, 21, 22, 23, 24, 25, 26, 27, 28, 29,
		30, 31, 32, 33, 34, 35, 36, 37, 38, 39,
		40, 41, 42, 43, 44, 45, 46, 47, 48, 49,
		50, 51, 52, 53, 54, 55, 56, 57, 58, 59,
		60, 61, 62, 63, 64, 65, 66, 67, 68, 69,
		70, 71, 72, 73, 74, 75, 76, 77, 78, 79,
		80, 81, 82, 83, 84, 85, 86, 87, 88, 89,
		90, 91, 92, 93, 94, 95, 96, 97, 98, 99,
		100, 101, 102, 103, 104, 105, 106, 107, 108, 109,
		110, 111, 112, 113, 114, 115, 116, 117, 118, 119,
		120, 121, 122, 123, 124, 125, 126, 127, 128, 129,
		130, 131, 132, 133, 134, 135, 136, 137, 138, 139,
		140, 141, 142, 143, 144, 145, 146, 147, 148, 149,
		150, 151, 152, 153, 154, 155, 156, 157, 158, 159,
		160, 161, 162, 163, 164, 165, 166, 167, 168, 169,
		170, 171, 172, 173, 174, 175, 176, 177, 178, 179,
		180, 181, 182, 183, 184, 185, 186, 187, 188, 189,
		190, 191, 192, 193, 194, 195, 196, 197, 198, 199,
		200, 201, 202, 203, 204, 205, 206, 207, 208, 209,
		210, 211, 212, 213, 214, 215, 216, 217, 218, 219,
		220, 221, 222, 223, 224, 225, 226, 227, 228, 229,
		230, 231, 232, 233, 234, 235, 236, 237, 238, 239,
		240, 241, 242, 243, 244, 245, 246, 247, 248, 249,
		250, 251, 252, 253, 254, 255,
	}
	self.x = 85
	self.y = 170
}

func (self *State8_t) Replace(in []byte) []byte {
	for i := range in {
		self.x = (self.x + 1) % 256
		self.y = (self.y + int(self.state[self.x]+in[i]) + 1) % 256
		self.state[self.x], self.state[self.y] = self.state[self.y], self.state[self.x]
		in[i] = self.state[(self.x+self.y)%256]
	}
	return in
}

func ReplaceUint64(state *State8_t, in uint64) uint64 {
	var temp [8]byte
	binary.BigEndian.PutUint64(temp[:], in)
	state.Replace(temp[:])
	return binary.BigEndian.Uint64(temp[:])
}
