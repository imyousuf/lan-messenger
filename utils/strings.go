package utils

import (
	"regexp"
	"strings"
)

// https://www.w3.org/TR/html5/forms.html#valid-e-mail-address
const (
	emailRegex     = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
	ipAddressRegex = `^(([0-9]|[0-9][0-9]|[1][0-9]{2}|[2][1-4][0-9]|[2][5][1-5])[.]){3}([0-9]|[0-9][0-9]|[1][0-9]{2}|[2][1-4][0-9]|[2][5][1-5])$`
)

var (
	isAlphaNumeric          = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString
	isAlphaNumericWithSpace = regexp.MustCompile(`^[a-zA-Z0-9 ]+$`).MatchString
	isValidEmailString      = regexp.MustCompile(emailRegex).MatchString
	isValidIPAddress        = regexp.MustCompile(ipAddressRegex).MatchString
	isValidConnectionString = regexp.MustCompile(ipAddressRegex[:len(ipAddressRegex)-1] + "[:][0-9]+$").MatchString
)

// IsStringAlphaNumeric checks whether the passed String is a alpha numeric only
func IsStringAlphaNumeric(String string) bool {
	return isAlphaNumeric(String)
}

// IsStringAlphaNumericWithSpace checks whether the passed String is a alpha numeric
// with space allowed only
func IsStringAlphaNumericWithSpace(String string) bool {
	return isAlphaNumericWithSpace(String)
}

// IsStringValidEmailFormat checks for valid email format
func IsStringValidEmailFormat(String string) bool {
	return isValidEmailString(String)
}

// IsValidIPv4Address checks for valid IP Addresses
func IsValidIPv4Address(String string) bool {
	return isValidIPAddress(String)
}

// IsValidConnectionString checks for ip-v4-address:port format of the string
func IsValidConnectionString(String string) bool {
	return isValidConnectionString(String)
}

// IsStringEmpty checks whether the string is empty
func IsStringEmpty(String string) bool {
	return len(String) <= 0
}

// IsStringNotEmpty checks whether the string is not empty
func IsStringNotEmpty(String string) bool {
	return len(String) > 0
}

// IsStringBlank checks whether the string is empty or filled with whitespace only
func IsStringBlank(String string) bool {
	return IsStringEmpty(String) || IsStringEmpty(strings.TrimSpace(String))
}

// IsStringNotBlank checks whether the string is neither empty nor has only whitespace
func IsStringNotBlank(String string) bool {
	return IsStringNotEmpty(String) && IsStringNotEmpty(strings.TrimSpace(String))
}
