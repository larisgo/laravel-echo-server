package errors

type Error struct {
	message string
}

func NewError(message string) error {
	return &Error{message: message}
}

func (this *Error) Error() string {
	return this.message
}
