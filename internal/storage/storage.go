package storage

import "errors"

var (
	ErrComplaintNotFound          = errors.New("complaints not found")
	ErrLimitOneComplaintInOneHour = errors.New("there are limit one complaint in one hour")
)
