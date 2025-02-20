package constant

import (
	"net/http"

	commonErr "github.com/chaihaobo/gocommon/error"
)

var (
	Successful           = commonErr.ServiceError{Code: "0000000", Message: "successful"}
	ErrorBadRequest      = commonErr.ServiceError{Code: "0000001", Message: "bad request"}
	ErrSystemMalfunction = commonErr.ServiceError{Code: "9999999", Message: "system malfunction"}
)

var ServiceErrorCode2HTTPStatus = map[string]int{
	Successful.Code:           http.StatusOK,
	ErrorBadRequest.Code:      http.StatusBadRequest,
	ErrSystemMalfunction.Code: http.StatusInternalServerError,
}

// MergeServiceErrorCode2HTTPStatus merge the given map into ServiceErrorCode2HTTPStatus
func MergeServiceErrorCode2HTTPStatus(m map[string]int) {
	for k, v := range m {
		ServiceErrorCode2HTTPStatus[k] = v
	}
}
