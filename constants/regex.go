package constants

const (
	// regex for credit cards, mobile
	RegexMobile = "[0-9+]{10,21}"

	// regex for email, vpa
	RegexEmail =
		"[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*"

	RegexIpAddress = `(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`

	RegexName = "name = '[a-z A-Z]*'"

	RegexNameReplace = "name = '?'"

	RegexDefaultReplace = "?"
)
