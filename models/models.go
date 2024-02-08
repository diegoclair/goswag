package models

type ReturnType struct {
	StatusCode int
	Body       interface{}
	// example: map[jsonFieldName]fieldType{}
	OverrideStructFields map[string]interface{}
}

type Swagger interface {
	// Summary is used to define the summary of the route.
	Summary(summary string) Swagger

	// If not set, the default value will be the same as the summary.
	Description(description string) Swagger

	// The name of group will be used as default if it is not empty and the tags are not defined.
	Tags(tags ...string) Swagger

	// The default value is json.
	// If you want to add a different value, check the swag documentation to know what are the possible values.
	// swag docs: https://github.com/swaggo/swag#mime-types
	Accepts(accept ...string) Swagger

	// The default value is json.
	// If you want to add a different value, check the swag documentation to know what are the possible values.
	// swag docs: https://github.com/swaggo/swag#mime-types
	Produces(produce ...string) Swagger

	// Read is used to define the request body of the route.
	Read(data interface{}) Swagger

	// Returns is used to define the return of the route.
	// The first parameter is the status code.
	// The second parameter is the body of the response.
	// The third parameter is used to override the fields of the response body, it is is optional.
	// Example:
	// if you have a response body like this:
	//
	//	type ResponseBody struct {
	//		ID   string `json:"id"`
	//		Data interface{} `json:"data"`
	//	}
	//
	// the swagger will be generated with the data field as a string field.
	// if you want to override the data field and specify that it is a struct for example, you can do this:
	// OverrideStructFields: map[string]interface{}{"data": SomeStruct{}}
	// where the SomeStruct{} is the struct that you want to use to override the "data" field.
	//
	// It accepts generic structs as well, but only for the first struct, if you have more deep generic fields, it may not work.
	//
	// Example using generic struct:
	//
	//	type ResponseBody[T any] struct {
	//			Data T   `json:"data"`
	//	}
	//
	// Then you will set the body like this:
	//
	//	ReturnType {
	//		StatusCode: http.StatusOK,
	//		Body: ResponseBody[SomeStruct]{},
	//	}
	Returns(data []ReturnType) Swagger

	// QueryParam is used to define the query parameters of the route and if it is required or not.
	QueryParam(name, description, dataType string, required bool) Swagger

	// HeaderParam is used to define the header parameters of the route and if it is required or not.
	HeaderParam(name, description, dataType string, required bool) Swagger
}
