package apperrors

type ErrCode string

const (
	Unknown ErrCode = "U000"

	BadParam             ErrCode = "R001"
	ReqBodyDecodeFailed  ErrCode = "R002"
	ResponseEncodeFailed ErrCode = "R003"
	NotFound             ErrCode = "R004"
	RequestBodyTooLarge  ErrCode = "R005"

	DependencyUnavailable ErrCode = "D001"
	DataMappingFailed     ErrCode = "D006"
)

func (code ErrCode) Wrap(err error, message string) error {
	return &Error{ErrCode: string(code), Message: message, Err: err}
}
