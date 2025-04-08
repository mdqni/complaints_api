package storage

import "errors"

var (
	ErrComplaintNotFound          = errors.New("complaints not found")
	ErrCategoryNotFound           = errors.New("category not found")
	ErrLimitOneComplaintInOneHour = errors.New("there are limit one complaint in one hour")
	ErrHasRelatedRows             = errors.New("there are related rows")
)
