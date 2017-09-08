package profile

import (
	"fmt"

	"github.com/imyousuf/lan-messenger/utils"
)

func ExampleNewUserProfile() {
	panicHandler := func(r interface{}) {
		fmt.Println("As expected panic handled:", r)
	}
	utils.PanicableInvocation(func() {
		NewUserProfile("a", "a", "a")
	}, panicHandler)
	utils.PanicableInvocation(func() {
		NewUserProfile("", "a", "a@a.com")
	}, panicHandler)
	utils.PanicableInvocation(func() {
		NewUserProfile("a", "", "a@a.com")
	}, panicHandler)
	utils.PanicableInvocation(func() {
		NewUserProfile("a", "a", "")
	}, panicHandler)
	utils.PanicableInvocation(func() {
		NewUserProfile("a", "a)", "a@a.co")
	}, panicHandler)
	utils.PanicableInvocation(func() {
		NewUserProfile("a_", "a", "a@a.co")
	}, panicHandler)
	uProfile := NewUserProfile("p", "B", "b@e.co")
	fmt.Println(uProfile.GetDisplayName(), uProfile.GetEmail(), uProfile.GetUsername())
	// Output:
	// As expected panic handled: Email is not well formatted!
	// As expected panic handled: None of the user profile attributes are optional
	// As expected panic handled: None of the user profile attributes are optional
	// As expected panic handled: None of the user profile attributes are optional
	// As expected panic handled: Username and Display Name must be Alpha Numeric only
	// As expected panic handled: Username and Display Name must be Alpha Numeric only
	// B b@e.co p
}
