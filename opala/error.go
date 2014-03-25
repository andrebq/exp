package opala

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	AtlasIsFull = Error("atlas is full")
)
