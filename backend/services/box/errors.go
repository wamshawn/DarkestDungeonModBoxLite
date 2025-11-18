package box

import "errors"

var (
	ErrOpened        = errors.New("err_opened")
	ErrGetSettings   = errors.New("err_get_settings")
	ErrSetSettings   = errors.New("err_set_settings")
	ErrMkDir         = errors.New("err_mk_dir")
	ErrIsNotDir      = errors.New("err_not_directory")
	ErrIsNotEmptyDir = errors.New("err_not_empty_directory")
	ErrInvalidFile   = errors.New("err_invalid_file")
)
