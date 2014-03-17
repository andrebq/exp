package phygo

// Implemented by object to give a unique Id.
type Ider interface {
	Id() Id
}

type Id uint

func (v Id) Valid() bool {
	return v != InvalidId
}

const (
	InvalidId = Id(0)
)
