package utils

import (
	"testing"
)

func TestIsStringAlphaNumeric(t *testing.T) {
	successCases := []string{"A1", "aa", "11", "A1a", "AA"}
	for _, successCase := range successCases {
		if !IsStringAlphaNumeric(successCase) {
			t.Error("Should pass as Alphanumeric: ", successCase)
			t.Fail()
		}
	}
	failCases := []string{"A$a", "9(o", "99.8"}
	for _, failCase := range failCases {
		if IsStringAlphaNumeric(failCase) {
			t.Error("Should not pass as Alphanumeric: ", failCase)
			t.Fail()
		}
	}
}

func TestIsStringValidEmailFormat(t *testing.T) {
	successCases := []string{"aa@aa", "aa@a.com", "a+1@asd.co"}
	for _, successCase := range successCases {
		if !IsStringValidEmailFormat(successCase) {
			t.Error("Should pass as Email: ", successCase)
			t.Fail()
		}
	}
	failCases := []string{"A$a", "9(o"}
	for _, failCase := range failCases {
		if IsStringValidEmailFormat(failCase) {
			t.Error("Should not pass as Email: ", failCase)
			t.Fail()
		}
	}
}

func TestIsValidIPv4Address(t *testing.T) {
	successCases := []string{"255.255.255.255", "0.0.0.0", "00.00.00.00", "127.0.0.1", "192.168.1.101"}
	for _, successCase := range successCases {
		if !IsValidIPv4Address(successCase) {
			t.Error("Should pass as IP Address: ", successCase)
			t.Fail()
		}
	}
	failCases := []string{"A$a", "9(o", "000.000.000.000", "0.0.256.0", "0.0.0.256", "0.0.0."}
	for _, failCase := range failCases {
		if IsValidIPv4Address(failCase) {
			t.Error("Should not pass as IP Address: ", failCase)
			t.Fail()
		}
	}
}

func TestIsValidConnectionString(t *testing.T) {
	successCases := []string{"127.0.0.1:3000", "192.168.1.101:2"}
	for _, successCase := range successCases {
		if !IsValidConnectionString(successCase) {
			t.Error("Should pass as Connection String: ", successCase)
			t.Fail()
		}
	}
	failCases := []string{"A$a", "127.0.0.1:", "127.0.0.1", ":3000"}
	for _, failCase := range failCases {
		if IsValidConnectionString(failCase) {
			t.Error("Should not pass as Connection String: ", failCase)
			t.Fail()
		}
	}
}

func TestStringEmptiness(t *testing.T) {
	if IsStringEmpty("") == false {
		t.Error("Wrong empty string detection")
	}
	if IsStringEmpty(" ") || IsStringEmpty("a") {
		t.Error("Wrong detection of empty string though not empty")
	}
	if IsStringNotEmpty("") {
		t.Error("Wrong non-empty string detection")
	}
	if IsStringNotEmpty(" ") == false || IsStringNotEmpty("a") == false {
		t.Error("Wrong detection of non-empty string though not empty")
	}
}

func TestStringBlankness(t *testing.T) {
	if IsStringBlank("") == false || IsStringBlank(" ") == false || IsStringBlank(" \t\r\n") == false {
		t.Error("Wrong blank string detection")
	}
	if IsStringBlank("a") {
		t.Error("Wrong detection of blank string though not blank")
	}
	if IsStringNotBlank("") || IsStringNotBlank(" ") || IsStringNotBlank(" \t\r\n") {
		t.Error("Wrong non-blank string detection")
	}
	if IsStringNotBlank("a") == false {
		t.Error("Wrong detection of non-blank string though not blank")
	}
}
