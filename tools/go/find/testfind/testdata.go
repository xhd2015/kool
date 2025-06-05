package testdata

type TestData struct {
	TestField string
}

func test() {
	var _ = TestData{
		TestField: "OK",
	}
}
