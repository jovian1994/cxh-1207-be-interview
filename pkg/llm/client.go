package llm

type ILLMClient interface {
	Translate(lang, content string, targetLang string) (string, error)
}

type llmClient struct{}

func NewLLMClient() ILLMClient {
	return &llmClient{}
}

func (c *llmClient) Translate(lang, content string, targetLang string) (string, error) {
	return "", nil
}
