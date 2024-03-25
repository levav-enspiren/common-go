package errorfactoryhttp

import (
	"encoding/json"
	"fmt"
	"net/http"

	eh "github.com/levav-enspiren/common-go/errorhandling"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Error struct {
	error
	Code    int
	Message string
}

func (err *Error) Error() string {
	return fmt.Sprintf("http error: code = %d desc = %s", err.Code, err.Message)
}

type ErrorFactory struct {
	eh.ErrorFactoryI
}

func (factory *ErrorFactory) ConvertCode(errorCode eh.Code) int {
	// The error code of error factory is the same as HTTP error code
	if errorCode == 0 {
		errorCode = eh.CodeUnknown
	}
	return int(errorCode)
}

func (factory *ErrorFactory) New(errorCode eh.Code, message string) error {
	var httpCode int = factory.ConvertCode(errorCode)
	return &Error{
		Code:    httpCode,
		Message: message,
	}
}

func (factory *ErrorFactory) Newf(errorCode eh.Code, formatString string, args ...interface{}) error {
	var httpCode int = factory.ConvertCode(errorCode)
	// return status.Errorf(gECode, formatString, args...)
	return &Error{
		Code:    httpCode,
		Message: fmt.Sprintf(formatString, args...),
	}
}

func codeGrpc2Http(grpcCode codes.Code) (httpCode int) {
	switch grpcCode {
	case codes.OK:
		return http.StatusOK
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}

func ResponseError(rError error) (statusCode int, obj gin.H) {
	statusCode = http.StatusInternalServerError
	message := rError.Error()
	// check if it is gRPC error
	grpcStatus, ok := status.FromError(rError)
	if ok {
		statusCode = codeGrpc2Http(grpcStatus.Code())
		message = grpcStatus.Message()
	} else {
		// check if it is custom error
		ehError, ok := rError.(*Error)
		if ok {
			statusCode = ehError.Code
			message = ehError.Message
		}
	}
	// parse error json
	err := json.Unmarshal([]byte(message), &obj)
	if err == nil {
		// check required fields
		if _, ok := obj["code"]; !ok {
			obj["code"] = ""
		}
		if _, ok := obj["message"]; !ok {
			obj["message"] = message
		}
	} else {
		// fallback to message
		obj = gin.H{
			"code":    "",
			"message": message,
		}
	}
	return
}

func StandardErrorHandling(context *gin.Context, err error) {
	statusCode, obj := ResponseError(err)
	println("[Error]", statusCode, obj)
	context.JSON(statusCode, obj)
}

var MsgFlags = map[int]string{
	http.StatusOK:                  "Success",
	http.StatusBadRequest:          "Bad request",
	http.StatusForbidden:           "Forbidden",
	http.StatusInternalServerError: "Internal error",
	http.StatusUnauthorized:        "Unauthorized",
	http.StatusNotFound:            "Not Found",
}

func SimpleResponse(context *gin.Context, statusCode int) {
	message, ok := MsgFlags[statusCode]
	if !ok {
		message = ""
	}
	context.JSON(statusCode, gin.H{
		"code":    statusCode,
		"message": message,
	})
}
