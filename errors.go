package errors

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
)

var (
	ErrNoData = New("data not found")
)

type Error interface {
	// extends for error interface
	Error() string
	// extends for json custom interface,
	// MarshalJSON() ([]byte, error)

	// return error code
	Code() string

	// Set error reason and positoin
	//
	// call as :
	// err.As(reason)
	// or errors.As(err, reason)
	//
	// return it self
	As(arg ...interface{}) Error

	// Compare error code
	// call as :
	// err.Equal(anotherErr)
	// or errors.Equal(err)
	Equal(err error) bool
}

// Equal
// compare err1 and err2 is same in memory index,error data or key code in engine/Err.
//
// Spec
// if Error Data is format with Err, compare with key code to the another error.
//
// Param
// err1 -- error one which want to compare.
// err2 -- error two which want to compare.
//
// Return
// return ture is same, or return false.
func Equal(err1 error, err2 error) bool {
	return errEqual(err1, err2)
}

func errEqual(err1 error, err2 error) bool {
	// memory compare
	if err1 == err2 {
		return true
	}
	if err1 == nil || err2 == nil {
		return false
	}

	// check Err type
	// if they are Err type,using errImpl compare.
	eImpl1, ok1 := err1.(*errImpl)
	eImpl2, ok2 := err2.(*errImpl)
	if ok1 && ok2 {
		return eImpl1.Code() == eImpl2.Code()
	}

	// if they are standar error,
	// compare the Message data.
	eMsg1 := err1.Error()
	eMsg2 := err2.Error()
	if eMsg1 == eMsg2 {
		return true
	}
	return parse(eMsg1).Code() == parse(eMsg2).Code()
}

type ErrData struct {
	Code   string          `json:"code"`
	Reason [][]interface{} `json:"reason"`
	Where  []string        `json:"where"`
}

type errImpl struct {
	data ErrData
}

// New
// create an Err implement error interface.
//
// Param
// code -- code or msg for the error struct,it will be a key.
//
// Return
// return a new Err interface
func New(code string) Error {
	return &errImpl{
		ErrData{
			Code:   code,
			Reason: [][]interface{}{{"new"}},
			Where:  []string{caller(2)},
		},
	}
}

// ParseErr
// Parse a standar error to Err interface.
// if the parameter is belong to Err, do a value copy an return a new Err.
// or parse string with error.Error(),
// if the string have a json struct with Err.Error(),return the origin struct with a new Err.
// or using error.Error() to create a new Err.
//
// Spec
// in the two case before, it will keep the key same as origin.
// the location is not change in parsing.
//
// Param
// src -- any error who implement error interface.
//
// Return
// return a new Err interface.
func ParseErr(src error) Error {
	if src == nil {
		return nil
	}
	if e, ok := src.(*errImpl); ok {
		return e
	}
	return parse(src.Error())
}

func Parse(src string) Error {
	if len(src) == 0 {
		return nil
	}
	return parse(src)
}

// As
// Parse the error, and fix with reason,it can make a replenishment for a same error.
//
// Spec
// because the value of error is change, so that location of Where is changed.
//
// Param
// err -- any error interface
// reason -- a array reason,it will be append to the reason of parameter.
//
// Return
// return a New Err,but with a same key with param error.
func As(err error, reason ...interface{}) Error {
	if err == nil {
		return nil
	}
	e := ParseErr(err).(*errImpl)
	e.data.Reason = append(e.data.Reason, reason)
	e.data.Where = append(e.data.Where, caller(2))
	return e

}

func parse(src string) *errImpl {
	if len(src) == 0 {
		return nil
	}
	if src[:1] != "{" {
		return New(src).(*errImpl)
	}

	data := ErrData{}
	if err := json.Unmarshal([]byte(src), &data); err != nil {
		return New(src).(*errImpl)
	}
	return &errImpl{data}
}

// call for domain
func caller(depth int) string {
	at := ""
	pc, file, line, ok := runtime.Caller(depth)
	if !ok {
		at = "domain of caller is unknown"
	}
	me := runtime.FuncForPC(pc)
	if me == nil {
		at = "domain of call is unnamed"
	}

	fileFields := strings.Split(file, "/")
	if len(fileFields) < 1 {
		at = "domain of file is unnamed"
		return at
	}
	funcFields := strings.Split(me.Name(), "/")
	if len(fileFields) < 1 {
		at = "domain of func is unnamed"
		return at
	}

	fileName := strings.Join(fileFields[len(fileFields)-1:], "/")
	funcName := strings.Join(funcFields[len(funcFields)-1:], "/")

	at = fmt.Sprintf("%s(%s:%d)", funcName, fileName, line)
	return at
}

func (e *errImpl) Code() string {
	return e.data.Code
}

func (e *errImpl) Error() string {
	data, err := json.Marshal(e.data)
	if err != nil {
		return fmt.Sprintf("%v", e.data)
	}
	return string(data)
}

func (e *errImpl) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.data)
}

func (e *errImpl) Equal(l error) bool {
	return errEqual(e, l)
}

func (e *errImpl) As(reason ...interface{}) Error {
	e.data.Reason = append(e.data.Reason, reason)
	e.data.Where = append(e.data.Where, caller(2))
	return e
}
