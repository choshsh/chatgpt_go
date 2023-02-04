package chatgpt_go

const (
	OpenAiHost    = "https://api.openai.com/v1"
	OpenAiTimeout = 60
)

const (
	EndpointCompletion = OpenAiHost + "/completions"
	EndpointModels     = OpenAiHost + "/models"
)

const (
	EofText = "data: [DONE]"
)
