package errs

import "fmt"

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string { return e.Message }

func New(code int, format string, a ...any) *AppError {
	return &AppError{Code: code, Message: fmt.Sprintf(format, a...)}
}

const (
	CodeOK                = 0
	CodeInvalidAPIToken   = 1001
	CodeInvalidParam      = 1002
	CodeInvalidIPv4       = 1003
	CodeDomainNotFound    = 1004
	CodeDomainUnavailable = 1005
	CodeNoAvailableDomain = 1006
	CodeSubdomainExists   = 1007
	CodeFQDNNotFound      = 1008
	CodeRecordNotFound    = 1009
	CodeIPNoBindings      = 1010
	CodeDomainAlreadyOff  = 1011
	CodeDomainAlreadyOn   = 1012
	CodeSyncFailed        = 1013
	CodeRateLimited       = 1014
	CodeProviderError     = 1015
	CodeDatabaseError     = 1016
	CodeInternal          = 1017
	CodeDomainConflict    = 1018
	CodeUniqueConflict    = 1019
	CodeIPFQDNMismatch    = 1020
)
