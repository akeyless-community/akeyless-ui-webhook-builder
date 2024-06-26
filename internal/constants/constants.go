package constants

const (
	FieldInputUsername       = "inputUsername"
	FieldInputPassword       = "inputPassword"
	FieldNewPassword         = "newPassword"
	FieldRequiredStaticInput = "requiredStaticInput"
)

var RequiredFields = []string{
	FieldInputUsername,
	FieldInputPassword,
	FieldNewPassword,
}

var OptionalFields = []string{
	FieldRequiredStaticInput,
}
