package zerror

import (
	"fmt"
)

const (
	NonBizErrCode = 0
	NonBizErrMsg  = "internal server error"
)

type BaseError struct {
	Code int
	Msg  string
	Err  error
	*stack
}

func (e *BaseError) Error() string { return e.listMsg(0) }
func (e *BaseError) Unwrap() error { return e.Err }
func (e *BaseError) Cause() error  { return e.Err }
func (e *BaseError) clone() *BaseError {
	return &BaseError{
		Code:  e.Code,
		Msg:   e.Msg,
		Err:   e.Err,
		stack: e.stack,
	}
}
func (e *BaseError) listMsg(sept int) string {
	var msg = e.Msg
	if e.stack == nil {
		return msg
	}
	firstFrame := e.stackTrace()[0]
	if temp, ok := e.Err.(*BaseError); ok {
		msg = fmt.Sprintf("\n #%d %s %s %s ", sept, msg, firstFrame, temp.listMsg(sept+1))
	} else {
		errMsg := "nil"
		if e.Err != nil {
			errMsg = e.Err.Error()
		}
		msg = fmt.Sprintf("\n #%d %s %s \n #e %s ",
			sept, msg, firstFrame, errMsg)
	}
	return msg
}

//New 创建一个业务异常,打印时日志级别为 warn,不为 error.
//配合接口返回值包装通用方法,会在最后返回时将 code,msg 注入到返回结构中
func New(code int, msg string) error {
	return &BaseError{
		Code:  code,
		Msg:   msg,
		Err:   nil,
		stack: nil,
	}
}

//Wrap 包装一个错误消息,打印日志时,会逐级打印错误信息
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return &BaseError{
		Msg:   msg,
		Err:   err,
		stack: callers(),
	}
}

//WrapCode 包装错误,并且带有错误码.
//包装后此错误会被认为是业务错误,打印日志时,会逐级打印 warn 级别错误信息
func WrapCode(err error, code int, msg string) error {
	return &BaseError{
		Code:  code,
		Msg:   msg,
		Err:   err,
		stack: callers(),
	}
}

//Verb 拿到第一个业务错误
//如果没有业务错误信息,则返回普通错误
func Verb(e error) (int, string) {
	errCode := NonBizErrCode
	errMsg := NonBizErrMsg

	for {
		temp, ok := e.(*BaseError)
		if !ok {
			break
		}
		if temp == nil {
			break
		}
		if temp.Code != 0 {
			errCode = temp.Code
			errMsg = temp.Msg
			break
		}
		e = temp.Err
	}
	return errCode, errMsg
}
