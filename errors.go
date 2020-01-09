package errors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Error struct {
	Code Code  `json:"code,omitempty"`
	Op   Op    `json:"op,omitempty"`
	Kind Kind  `json:"kind,omitempty"`
	Err  error `json:"err,omitempty"`
	*stack
}

type Code int
type Op string
type Kind uint8

const (
	Other             Kind = iota // Unclassified error. This value is not printed in the error message.
	Invalid                       // Invalid operation, or the request is in some way invalid i.e. has failed a validation.
	Permission                    // Permission denied.
	IO                            // External I/O error e.g. disk or network.
	Exist                         // Item already exists.
	NotExist                      // Item does not exist.
	Timeout                       // Timeout has occured e.g. transient network error.
	Database                      // A database error has occured e.g. deadlock.
	Encoding                      // An encoding error has occured e.g. writing an HTTP response.
	Decoding                      // A decoding error has occured e.g. decoding an HTTP request body.
	HTTP                          // An HTTP error not related to network issues e.g. generating a request.
	DuplicateKey                  // A duplicate key error from the database.
	Canceled                      // A request was canceled using the provided context.
	Unimplemented                 // The feature has not been implemented
	UnsupportedSyntax             // The syntax is not supported e.g. in a parser.
)

const (
	CodeBadRequest   = Code(http.StatusBadRequest)          // Request was malformed.
	CodeServerError  = Code(http.StatusInternalServerError) // Unspecified server error.
	CodeInvalid      = Code(http.StatusUnprocessableEntity) // A business logic related error.
	CodeUnauthorized = Code(http.StatusUnauthorized)        // The user has not been authenticated.
	CodeForbidden    = Code(http.StatusForbidden)           // The user is not authorized.
	CodeNotFound     = Code(http.StatusNotFound)            // The resource was not found.
)

func E(args ...interface{}) error {
	if len(args) == 0 {
		panic("No arguments provided to errors.E")
	}

	e := &Error{
		stack: callingLocations(6, 3),
	}

	for _, arg := range args {
		switch arg := arg.(type) {
		case Code:
			e.Code = arg
		case Op:
			e.Op = arg
		case string:
			e.Err = Str(arg) //convert string to basic error
		case Kind:
			e.Kind = arg
		case *Error:
			// Make a copy
			if arg == nil {
				return nil
			}
			copy := *arg
			e.Err = &copy
		case error:
			if arg == nil {
				return nil
			}
			e.Err = arg
		case nil:
			return nil
		default:
			return Errorf("unknown type %T, value %v in error call", arg, arg) //just return a basic error
		}
	}

	if e.Err == nil {
		e.Err = Str("")
	}

	prev, ok := e.Err.(*Error)
	if !ok {
		return e
	}

	// if we make it to here then the error we are wrapping was already one of ours. We must clear out the values so we don't have duplication, we just keep what's different
	if prev.Code == e.Code {
		prev.Code = 0
	}
	if prev.Kind == e.Kind {
		prev.Kind = Other
	}
	// If this error has Kind unset or Other, pull up the inner one.
	if e.Kind == Other {
		e.Kind = prev.Kind
		prev.Kind = Other
	}

	return e
}

func (e *Error) Error() string {
	b := new(bytes.Buffer)

	if e.Code != 0 {
		prepend(b, ": ")
		b.WriteString("code ")
	}

	if e.Kind != 0 {
		prepend(b, ": ")
		b.WriteString(e.Kind.String())
	}

	if e.Op != "" {
		prepend(b, ": ")
		b.WriteString(string(e.Op))
	}

	if e.Err != nil {
		prepend(b, "\n")
		b.WriteString(e.Err.Error())
	}

	if b.Len() == 0 {
		return "No error"
	}

	return b.String()
}

func IsNoRows(err error) bool {
	return strings.Contains(err.Error(), "now rows in result set")
}

//function just implements the json.Marshaller interface
func (e *Error) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer

	b.WriteString(`{"code":`)
	b.WriteString(strconv.Itoa(int(e.Code)))

	b.WriteString(`,"op":"`)
	b.WriteString(string(e.Op))

	b.WriteString(`","kind":`)
	b.WriteString(strconv.Itoa(int(e.Kind)))

	b.WriteString(`,"err":`)
	b.WriteString(fmt.Sprintf("%q", e.Err.Error()))

	b.WriteString(`,"stack":`)
	buf, err := json.Marshal(e.stack)
	if err != nil {
		return nil, err
	}
	b.Write(buf)

	b.WriteString("}")

	return b.Bytes(), nil
}

func prepend(b *bytes.Buffer, s string) {
	if b.Len() == 0 {
		return //no need to add anything
	}
	b.WriteString(s)
}

func (k Kind) String() string {
	switch k {
	case Exist:
		return "item already exists"
	case NotExist:
		return "item does not exist"
	case Invalid:
		return "invalid operation"
	case Permission:
		return "permission denied"
	case IO:
		return "I/O error"
	case Timeout:
		return "timeout error"
	case Database:
		return "database error"
	case Encoding:
		return "encoding error"
	case Decoding:
		return "decoding error"
	case HTTP:
		return "HTTP error"
	case Other:
		return "other error"
	case DuplicateKey:
		return "duplicate key error"
	case Canceled:
		return "request canceled"
	case Unimplemented:
		return "unimplemented"
	}
	return "unknown"
}

func Str(t string) error {
	return &stringError{t}
}

type stringError struct { //all this does is satisfy the error interface so that we can return a string an error
	s string
}

func (e *stringError) Error() string { //satisfy the interface
	return e.s
}

func Errorf(format string, args ...interface{}) error {
	return &stringError{fmt.Sprintf(format, args...)}
}
