package errorhandling

type Code uint32

// Error code system with gRPC-like naming and HTTP-like value
var CodeOk Code = 200 // This is not for error. Properly a reference for return code
var CodeInvalidArgument Code = 400
var CodeUnauthenticated Code = 401
var CodePermissionDenied Code = 403
var CodeDeadlineExceeded Code = 408
var CodeAlreadyExists Code = 409
var CodeUnknown Code = 500
var CodeUnimplemented Code = 501

type ErrorFactoryI interface {
	New(errorCode Code, message string) error
	Newf(errorCode Code, formatString string, args ...interface{}) error
}

var ErrorFactory ErrorFactoryI

func SetErrorFactory(ef ErrorFactoryI) {
	ErrorFactory = ef
}

// Helper functions
func New(errorCode Code, message string) error {
	return ErrorFactory.New(errorCode, message)
}

func Newf(errorCode Code, formatString string, args ...interface{}) error {
	return ErrorFactory.Newf(errorCode, formatString, args...)
}
