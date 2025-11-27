package archives

import (
	"errors"
	"fmt"
)

var (
	ErrPasswordRequired = errors.New("password required")
	ErrPasswordInvalid  = errors.New("password invalid")
)

type PasswordFailed struct {
	Filename         string
	PasswordRequired bool
	PasswordInvalid  bool
}

func (e PasswordFailed) Error() string {
	return e.String()
}

func (e PasswordFailed) String() string {
	if e.PasswordInvalid {
		return fmt.Sprintf("Password invalid: %s", e.Filename)
	}
	return fmt.Sprintf("Password required: %s", e.Filename)
}

type FileError struct {
	Filename string
	Err      error
}

func (e FileError) Error() string {
	return fmt.Sprintf("%s: %v", e.Filename, e.Err)
}

func (e FileError) Unwrap() error { return e.Err }

func IsPasswordFailed(err error) (errs []PasswordFailed, ok bool) {
	if err == nil {
		return
	}

	if joinErr, isJoined := err.(interface {
		Unwrap() []error
	}); isJoined {
		for _, e := range joinErr.Unwrap() {
			if subs, subsOk := IsPasswordFailed(e); subsOk && len(subs) > 0 {
				errs = append(errs, subs...)
			}
		}
		ok = len(errs) > 0
		return
	}
	fErr := FileError{}
	if errors.As(err, &fErr) {
		e := PasswordFailed{
			Filename:         fErr.Filename,
			PasswordRequired: false,
			PasswordInvalid:  false,
		}
		if errors.Is(fErr.Err, ErrPasswordRequired) {
			e.PasswordRequired = true
		} else if errors.Is(fErr.Err, ErrPasswordInvalid) {
			e.PasswordInvalid = true
		}
		errs = append(errs, e)
		ok = true
	}
	return
}
