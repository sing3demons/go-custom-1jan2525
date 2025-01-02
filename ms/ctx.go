package ms

type IContext interface {
	Log(message string)
	Param(name string) string
	Query(name string) string 
	ReadInput(data interface{}) error
	Response(responseCode int, responseData interface{})
}
