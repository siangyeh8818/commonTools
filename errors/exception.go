package errors

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/proto"
	"github.com/jackc/pgconn"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// exception for custom error
type exception struct {
	Code     string
	Status   int
	Message  string
	Details  map[string]interface{}
	GRPCCode codes.Code
	_e       error
}

// ErrorView for client
type ErrorView struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Error implement error interface
func (e *exception) Error() string {
	var b strings.Builder
	_, _ = b.WriteRune('[')
	_, _ = b.WriteString(e.Code)
	_, _ = b.WriteRune(']')
	_, _ = b.WriteRune(' ')
	if e._e != nil {
		_, _ = b.WriteString(e._e.Error())
	} else {
		_, _ = b.WriteString(e.Message)
	}
	return b.String()
}

// Is target err equal this error
func Is(err error, target error) bool {
	causeErr, ok := errors.Cause(err).(*exception)
	if !ok {
		return false
	}
	causeTarget, ok := errors.Cause(target).(*exception)
	if !ok {
		return false
	}
	return causeErr.Code == causeTarget.Code
}

// Is target err equal this error
func (e *exception) Is(err error) bool {
	causeErr, ok := errors.Cause(err).(*exception)
	if !ok {
		return false
	}
	return e.Code == causeErr.Code
}

// WithErrors 使用訂好的errors code 與訊息,如果未定義message 顯示對應的http status描述
func WithErrors(err error) error {
	if err == nil {
		return nil
	}
	causeErr := errors.Cause(err)
	_err, ok := causeErr.(*exception)
	if !ok {
		return WithStack(&exception{
			Status:  ErrInternal.Status,
			Code:    ErrInternal.Code,
			Message: http.StatusText(ErrInternal.Status),
		})
	}
	return WithStack(&exception{
		Status:  _err.Status,
		Code:    _err.Code,
		Message: _err.Message,
	})
}

// Wrap returns an error annotating err with a stack trace
// at the point Wrap is called, and the supplied Message.
// If err is nil, Wrap returns nil.
func Wrap(err error, msg string) error {
	_w := errors.Wrap(err, msg)
	_e, ok := err.(*exception)
	if !ok {
		return _w
	}
	_err := *_e
	_err._e = _w
	return &_err

}

// Wrapf WithMessage annotates err with a new Message.
// If err is nil, WithMessage returns nil.
func Wrapf(err error, format string, args ...interface{}) error {
	_w := errors.Wrapf(err, format, args...)
	_e, ok := err.(*exception)
	if !ok {
		return _w
	}
	_err := *_e
	_err._e = _w
	return &_err
}

// NewWithMessage 抽換錯誤訊息
// 未定義的錯誤會被視為 ErrInternalError 類型
func NewWithMessage(err error, message string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	causeErr := errors.Cause(err)
	_err, ok := causeErr.(*exception)
	if !ok {
		return WithStack(&exception{
			Status:   ErrInternal.Status,
			Code:     ErrInternal.Code,
			Message:  ErrInternal.Message,
			GRPCCode: ErrInternal.GRPCCode,
		})
	}
	err = &exception{
		Status:   _err.Status,
		Code:     _err.Code,
		Message:  message,
		GRPCCode: _err.GRPCCode,
	}
	var msg string
	for i := 0; i < len(args); i++ {
		msg = msg + "%+v"
	}
	return Wrapf(err, msg, args...)
}

// WithStack is as the proxy for github.com/pkg/errors.WithStack func.
// func WithStack(err error) error {
// 	return errors.WithStack(err)
// }
var WithStack = errors.WithStack

func GetHttpError(err *exception) ErrorView {
	return ErrorView{
		Message: err.Message,
		Code:    err.Code,
		Details: err.Details,
	}
}

// SetDetails set details as you wish =)
func (e *exception) SetDetails(details map[string]interface{}) {
	e.Details = details
	return
}

// ConvertMySQLError convert mysql error
func ConvertMySQLError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrResourceNotFound
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrConflict
	}
	return ErrInternal
}

// ConvertPostgresError convert postgres error
func ConvertPostgresError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrResourceNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrConflict
		}
	}

	return ErrInternal
}

// IsRedisResourceNotFound is this error is redis nil
func IsRedisNil(err error) bool {
	return errors.Is(err, redis.Nil)
}

// ToRestfulView for http view
func ToRestfulView(target error) *ErrorView {
	if target == nil {
		return nil
	}
	err, ok := errors.Cause(target).(*exception)
	if !ok {
		return &ErrorView{
			Code:    ErrInternal.Code,
			Message: http.StatusText(ErrInternal.Status),
		}
	}
	return &ErrorView{
		Code:    err.Code,
		Message: err.Message,
		Details: err.Details,
	}
}

// ToRestfulView for http view
func ToWebsocketView(target error) (code, msg string, data []byte) {
	if target == nil {
		return "00000", "", []byte{}
	}
	err, ok := errors.Cause(target).(*exception)
	if !ok {
		return ErrInternal.Code, http.StatusText(ErrInternal.Status), []byte{}
	}

	for _, v := range err.Details {
		var pms proto.Message
		if pms, ok = v.(proto.Message); !ok {
			continue
		}
		b, err := proto.Marshal(pms)
		if err != nil {
			continue
		}
		data = b
	}
	return err.Code, err.Message, data
}

//ConvertHttpErr Convert  grpc error to _error
func ConvertHttpErr(err error) error {
	if err == nil {
		return nil
	}
	s := status.Convert(err)
	if s == nil {
		return ErrInternal
	}
	interErr := exception{}
	jerr := json.Unmarshal([]byte(s.Message()), &interErr)
	if jerr != nil {
		return switchCode(s)
	}
	return WithStack(&interErr)
}

func switchCode(s *status.Status) error {
	httperr := ErrInternal
	switch s.Code() {
	case Unknown:
		httperr = ErrInternal
	case InvalidArgument:
		httperr = ErrInvalidInput
	case NotFound:
		httperr = ErrResourceNotFound
	case AlreadyExists:
		httperr = ErrConflict
	case PermissionDenied:
		httperr = ErrNotAllowed
	case Unauthenticated:
		httperr = ErrUnauthorized
	case OutOfRange:
		httperr = ErrInvalidInput
	case Internal:
		httperr = ErrInternal
	case DataLoss:
		httperr = ErrInternal
	}
	httperr.Message = s.Message()
	return WithStack(httperr)
}

//ConvertProtoErr Convert _error to grpc error
func ConvertProtoErr(err error) error {
	if err == nil {
		return nil
	}
	causeErr := errors.Cause(err)
	_err, ok := causeErr.(*exception)
	if !ok {
		return status.Error(ErrInternal.GRPCCode, err.Error())
	}
	b, _ := json.Marshal(_err)
	return status.Error(_err.GRPCCode, string(b))
}
