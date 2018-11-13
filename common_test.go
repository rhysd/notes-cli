package notes

// Test utilities

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
