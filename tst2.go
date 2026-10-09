//
// ternary search tree
//

package tst

import "math"

type node2_t[Value_t any] struct {
	hi_kid int
	eq_kid int
	lo_kid int
	value  *Mapped_t[Value_t]
	key    rune
}

type Tree2_t[Value_t any] struct {
	root []node2_t[Value_t]
}

type Cursor2_t[Value_t any] struct {
	root []node2_t[Value_t]
	cur  int
}

func (self *Tree2_t[Value_t]) Add(in string, value Value_t) {
	cur := 0
	last := math.MaxInt
	for _, key := range in {
		for cur < len(self.root) && key != self.root[cur].key {
			if key < self.root[cur].key {
				if self.root[cur].lo_kid == math.MaxInt {
					self.root[cur].lo_kid = len(self.root)
				}
				cur = self.root[cur].lo_kid
			} else {
				if self.root[cur].hi_kid == math.MaxInt {
					self.root[cur].hi_kid = len(self.root)
				}
				cur = self.root[cur].hi_kid
			}
		}
		if cur >= len(self.root) {
			cur = len(self.root)
			if last != math.MaxInt && self.root[last].eq_kid == math.MaxInt {
				self.root[last].eq_kid = cur
			}
			self.root = append(self.root, node2_t[Value_t]{key: key, eq_kid: math.MaxInt, hi_kid: math.MaxInt, lo_kid: math.MaxInt})
		}
		last = cur
		cur = self.root[cur].eq_kid
	}
	if last != math.MaxInt {
		self.root[last].value = &Mapped_t[Value_t]{value: value}
	}
}

func (self *Tree2_t[Value_t]) Cursor() (c *Cursor2_t[Value_t]) {
	return &Cursor2_t[Value_t]{root: self.root}
}

func (self *Cursor2_t[Value_t]) Fetch(key rune) (value Value_t, ok bool, next bool) {
	if len(self.root) == 0 {
		return
	}
	for self.cur < len(self.root) && key != self.root[self.cur].key {
		if key < self.root[self.cur].key {
			self.cur = self.root[self.cur].lo_kid
		} else {
			self.cur = self.root[self.cur].hi_kid
		}
	}
	if self.cur == math.MaxInt {
		return
	}
	if ok = self.root[self.cur].value != nil; ok {
		value = self.root[self.cur].value.value
	}
	self.cur = self.root[self.cur].eq_kid
	next = true
	return
}

func (self *Tree2_t[Value_t]) Search(in string) (value Value_t, found int) {
	c := self.Cursor()
	for _, symbol := range in {
		temp, okvalue, next := c.Fetch(symbol)
		if okvalue {
			found++
			value = temp
		}
		if next == false {
			return
		}
	}
	return
}
