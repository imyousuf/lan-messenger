package utils

import (
	"fmt"
	"testing"
)

func TestPanicableInvocation(t *testing.T) {
	plainMainTask := func() {
		// Do nothing
	}
	unexpectedPanicHandler := func(reason interface{}) {
		t.Error("Did not expect any panic:", reason)
	}
	PanicableInvocation(plainMainTask, unexpectedPanicHandler)
	expectedReason := "Random panic"
	panicTask := func() {
		panic(expectedReason)
	}
	var receivedReason string
	expectedPanicHandler := func(reason interface{}) {
		receivedReason = reason.(string)
	}
	PanicableInvocation(panicTask, expectedPanicHandler)
	if expectedReason != receivedReason {
		t.Error("Did not receive panic:", receivedReason)
	}
}

func ExamplePanicableInvocation() {
	plainMainTask := func() {
		fmt.Println("Execute")
	}
	unexpectedPanicHandler := func(reason interface{}) {
	}
	PanicableInvocation(plainMainTask, unexpectedPanicHandler)
	expectedReason := "Random panic"
	panicTask := func() {
		panic(expectedReason)
	}
	expectedPanicHandler := func(reason interface{}) {
		fmt.Println("Recovered from:", expectedReason)
	}
	PanicableInvocation(panicTask, expectedPanicHandler)
	// Output:
	// Execute
	// Recovered from: Random panic
}
