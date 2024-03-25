package errorfactorygrpc

import (
	"gitea.greatics.net/common-go/errorhandling"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorFactory struct {
	errorhandling.ErrorFactoryI
}

func (factory *ErrorFactory) ConvertCode(errorCode errorhandling.Code) codes.Code {
	var gECode codes.Code = codes.Unknown
	switch errorCode {
	case errorhandling.CodeOk:
		gECode = codes.OK
	case errorhandling.CodeInvalidArgument:
		gECode = codes.InvalidArgument
	case errorhandling.CodePermissionDenied:
		gECode = codes.PermissionDenied
	case errorhandling.CodeDeadlineExceeded:
		gECode = codes.DeadlineExceeded
	case errorhandling.CodeAlreadyExists:
		gECode = codes.AlreadyExists
	case errorhandling.CodeUnimplemented:
		gECode = codes.Unimplemented
	case errorhandling.CodeUnauthenticated:
		gECode = codes.Unauthenticated
	}
	return gECode
}

func (factory *ErrorFactory) New(errorCode errorhandling.Code, message string) error {
	var gECode codes.Code = factory.ConvertCode(errorCode)
	return status.Error(gECode, message)
}

func (factory *ErrorFactory) Newf(errorCode errorhandling.Code, formatString string, args ...interface{}) error {
	var gECode codes.Code = factory.ConvertCode(errorCode)
	return status.Errorf(gECode, formatString, args...)
}
