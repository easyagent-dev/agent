package agent

type Agent struct {
	Name         string
	Description  string
	Instructions string
	Tools        []ModelTool
	Callback     Callback
}
