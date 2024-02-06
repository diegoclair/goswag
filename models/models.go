package models

type ReturnType struct {
	StatusCode int
	Body       interface{}
	// example: map[jsonFieldName]fieldType{}
	OverrideStructFields map[string]interface{}
}
