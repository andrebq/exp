package opala

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	AtlasIsFull         = Error("atlas is full")
	CountMustBePositive = Error("count must be positive")
	GlfwUnableToInit    = Error("unable to initialize glfw. check the log")
)
