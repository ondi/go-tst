//
// ternary search tree
//

package tst

type node1_t[Value_t any] struct {
	eq_kid *node1_t[Value_t]
	hi_kid *node1_t[Value_t]
	lo_kid *node1_t[Value_t]
	value  *Mapped_t[Value_t]
	key    rune
}

type Tree1_t[Value_t any] struct {
	root *node1_t[Value_t]
}

func (self *Tree1_t[Value_t]) Root() *node1_t[Value_t] {
	return self.root
}

func (self *Tree1_t[Value_t]) Add(in string, value Value_t) {
	cur := &self.root
	var last **node1_t[Value_t]
	for _, key := range in {
		for *cur != nil && key != (*cur).key {
			if key < (*cur).key {
				cur = &(*cur).lo_kid
			} else {
				cur = &(*cur).hi_kid
			}
		}
		if *cur == nil {
			*cur = &node1_t[Value_t]{key: key}
		}
		last = cur
		cur = &(*cur).eq_kid
	}
	if last != nil {
		(*last).value = &Mapped_t[Value_t]{value: value}
	}
}

func Fetch[Value_t any](in *node1_t[Value_t], key rune) (next *node1_t[Value_t], res *node1_t[Value_t], ok bool) {
	res = in
	for res != nil && key != res.key {
		if key < res.key {
			res = res.lo_kid
		} else {
			res = res.hi_kid
		}
	}
	if res == nil {
		return
	}
	next = res.eq_kid
	ok = res.value != nil
	return
}

func (self *Tree1_t[Value_t]) Search(in string) (value Value_t, found int) {
	var okfetch bool
	var temp *node1_t[Value_t]
	next := self.Root()
	for _, symbol := range in {
		next, temp, okfetch = Fetch(next, symbol)
		if okfetch {
			found++
			value = temp.value.value
		}
		if next == nil {
			return
		}
	}
	return
}
