package examples

type Reply struct {
	Reply string `json:"reply" jsonschema:"required,reply=You final reply to user"`
}
