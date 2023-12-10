package orm

import "errors"

var (
	ErrMustNotNil         = errors.New("data model must not be nil")
	ErrMustBeStruct       = errors.New("data value must be type struct")
	ErrFieldNotFound      = errors.New("field not found")
	ErrTagValueNotFound   = errors.New("tag value not found")
	ErrNotIdentifyFkField = errors.New("not identify fk field on tag")
	ErrRegistryNotFound   = errors.New("registry not found")
)
