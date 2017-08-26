package utils

// PanicableInvocation provides a boilerplate code to invoke functions
// where panic is expected in case of errors. This function provides a
// similar structure to a try/catch block
func PanicableInvocation(mainTask func(), panicHandler func(interface{})) {
	defer func() {
		if r := recover(); r != nil {
			panicHandler(r)
		}
	}()
	mainTask()
}
