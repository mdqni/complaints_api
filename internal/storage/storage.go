package storage

import "errors"

var (
	ErrCategoryNotFound = errors.New("categories not found")
	ErrHasRelatedRows   = errors.New("there are related rows")
	ErrCreateCategory   = errors.New("failed to create categories")
	//----------------------
	ErrCreateComplaint            = errors.New("failed to create categories")
	ErrLimitOneComplaintInOneHour = errors.New("there are limit one complaint in one hour")
	ErrComplaintNotFound          = errors.New("complaints not found")
	//----------------------
	ErrDBConnection = errors.New("database connection error")
	ErrScanFailure  = errors.New("failed to scan row from DB")
)
