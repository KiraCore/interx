package log

// Recovery function to handle panics
func RecoverFromPanic() {
	if r := recover(); r != nil {
		CustomLogger().Error("Application crashed",
			"recovered pnic", r,
		)
	}
}
