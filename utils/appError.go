package utils

var ErrForbidden = AppError{"Forbidden", -403, nil}
var ErrInternalServer = AppError{"Internal server error", -500, nil}

type AppError struct {
	Message string
	Code    int
	Data    interface{}
}

func (e AppError) Error() string {
	return e.Message
}
