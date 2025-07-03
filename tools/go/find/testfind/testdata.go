package testdata

type TestData struct {
	TestField string
}

type TestDataRenamed TestData

type TestDataAlias = TestData

func testLiteral() {
	var _ = TestData{
		TestField: "OK",
	}
}

func testPointer() {
	t := &TestData{}

	t.TestField = "x"
}

func testRenamed() {
	t := TestDataRenamed{}

	t.TestField = "x"
}

func testAlias() {
	t := TestDataAlias{}

	t.TestField = "x"
}
