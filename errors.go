package cedar

import "errors"

// defines Error type
var (
	ErrInvalidDataType = errors.New("cedar: invalid datatype")
	ErrInvalidValue    = errors.New("cedar: invalid value")
	ErrInvalidKey      = errors.New("cedar: invalid key")
	ErrNoPath          = errors.New("cedar: no path")
	ErrNoValue         = errors.New("cedar: no value")
	ErrTooLarge        = errors.New("acmatcher: Tool Large for grow")
	ErrNotCompile      = errors.New("acmatcher: matcher must be compiled")
	ErrAlreadyCompiled = errors.New("acmatcher: matcher already compile")
)
