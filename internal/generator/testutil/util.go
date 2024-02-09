package testutil

type StructGeneric[T any] struct {
	Body T
}

type TestGeneric struct {
	Name string
}

type OverrideStruct struct {
	Body interface{} ` json:"body" `
}
