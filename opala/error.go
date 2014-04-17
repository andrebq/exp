package opala

import (
	"fmt"
	"github.com/go-gl/gl"
	"github.com/go-gl/glu"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

func checkGlError(expeced ...gl.GLenum) error {
	gle := gl.GetError()
	if gle != gl.NO_ERROR {
		str, err := glu.ErrorString(gle)
		if err != nil {
			str = err.Error()
		}
		for _, v := range expeced {
			if v == gle {
				return Error(fmt.Sprintf("[gl][expected] %v: %v", gle, str))
			}
		}
		return Error(fmt.Sprintf("[gl][unexpected] %v: %v", gle, str))
	}
	return nil
}

const (
	AtlasIsFull         = Error("atlas is full")
	CountMustBePositive = Error("count must be positive")
	GlfwUnableToInit    = Error("unable to initialize glfw. check the log")
)
