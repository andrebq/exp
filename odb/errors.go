package odb

import (
	"fmt"
)

type Error struct {
	code uint
	err  string
}

const (
	InvalidIndexFind    = 1
	UnableToReadStorage = 2
	NoIndexProvided     = 4
	ObjectFromOtherDB   = 8
	InvalidType         = 16
)

var (
	errInvalidIndexFind = Error{
		InvalidIndexFind,
		"invalid input data for index find"}
	errUnableToReadStorage = Error{
		UnableToReadStorage,
		"unable to read the storage information",
	}
	errNoIndexProvided = Error{
		NoIndexProvided,
		"no index provided for the operation",
	}
	errObjectFromOtherDB = Error{
		ObjectFromOtherDB,
		"the object belgons to another db and can be manipulated by this one",
	}
	errInvalidType = Error{
		InvalidType,
		"unable to read/write the given type",
	}
)

func newError(code uint, message string, data ...interface{}) Error {
	return Error{code, fmt.Sprintf(message, data...)}
}

func (e Error) Code() uint {
	return e.code
}

func (e Error) Error() string {
	return e.err
}
