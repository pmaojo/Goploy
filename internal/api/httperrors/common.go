package httperrors

import (
	"net/http"

	"github.com/pmaojo/goploy/internal/types"
)

var (
	ErrBadRequestZeroFileSize = NewHTTPError(http.StatusBadRequest, types.PublicHTTPErrorTypeZEROFILESIZE, "File size of 0 is not supported.")
)
