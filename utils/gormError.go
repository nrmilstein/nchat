package utils

type GormError struct {
	err error
	msg string
}

func (e GormError) Error() string {
	return e.msg
}

func (e GormError) Unwrap() error {
	return e.err
}

func NewGormError(e error) GormError {
	return GormError{
		err: e,
		msg: "Gorm error: " + e.Error(),
	}
}
