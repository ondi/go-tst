//
// ternary search tree
//

package tst

import "unicode/utf8"

type TernaryNode2_t struct {
	hi_kid int
	eq_kid int
	lo_kid int
	key rune
	value string	// prefix terminator
}

type TernaryTree2_t struct {
	root []TernaryNode2_t
}

type Cursor2_t struct {
	root []TernaryNode2_t
	cur int
}

func (self * TernaryTree2_t) Add(str string, value string) {
	last := -1
	var cur int
	if len(self.root) == 0 {
		cur = -1
	}
	for _, key := range str {
		for cur != -1 && key != self.root[cur].key {
			last = cur
			if key < self.root[cur].key {
				cur = self.root[cur].lo_kid
				if cur == -1 {
					self.root[last].lo_kid = len(self.root)
				}
			} else {
				cur = self.root[cur].hi_kid
				if cur == -1 {
					self.root[last].hi_kid = len(self.root)
				}
			}
		}
		if cur == -1 {
			cur = len(self.root)
			if last != -1 && self.root[last].eq_kid == -1 {
				self.root[last].eq_kid = cur
			}
			self.root = append(self.root, TernaryNode2_t{key: key, eq_kid: -1, hi_kid: -1, lo_kid: -1})
		}
		last = cur
		cur = self.root[cur].eq_kid
	}
	if last != -1 {
		self.root[last].value = value
	}
}

func (self * TernaryTree2_t) Cursor() (c * Cursor2_t) {
	c = &Cursor2_t{}
	c.root = self.root
	if len(self.root) == 0 {
		c.cur = -1
	}
	return
}

func (self * Cursor2_t) Fetch(key rune) (value string, next bool) {
	for self.cur != -1 && key != self.root[self.cur].key {
		if key < self.root[self.cur].key {
			self.cur = self.root[self.cur].lo_kid
		} else {
			self.cur = self.root[self.cur].hi_kid
		}
	}
	if self.cur == -1 {
		return value, false
	}
	if len(self.root[self.cur].value) > 0 {
		value = self.root[self.cur].value
	}
	self.cur = self.root[self.cur].eq_kid
	return value, self.cur != -1
}

func (self * TernaryTree2_t) Search(str string) (int, int, string, bool) {
	var found string
	var value string
	var next bool
	last := 0
	c := self.Cursor()
	for n, key := range str {
		value, next = c.Fetch(key)
		if len(value) > 0 {
			found = value
			_, size := utf8.DecodeRuneInString(str[n:])
			last = n + size
		}
		if next == false {
			return last, n, found, false
		}
	}
	return last, len(str), found, next
}
