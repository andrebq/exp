package chipmunk

// just some helper functions/types

type Freer interface {
	Free()
}

// releases the list of objects using
// a LIFO list
type freeStack []Freer

func newFreeStack() freeStack {
	return make(freeStack, 0)
}

func (fl *freeStack) push(toFree Freer) {
	*fl = append(*fl, toFree)
}

func (fl *freeStack) Free() {
	for i := len(*fl) - 1; i >= 0; i-- {
		(*fl)[i].Free()
		(*fl)[i] = nil
	}
	*fl = (*fl)[:0]
}
