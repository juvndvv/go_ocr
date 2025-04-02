package ai

type Factory struct {
}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) Create(aiType string) (Model, error) {
	switch aiType {
	case "deepseek-reasoner":
		return f.createDeepseekReasonerModel()
	case "deepseek-chat":
		return f.createDeepseekChatModel()
	}

	return nil, ErrUnsupportedAiType
}

func (f *Factory) createDeepseekReasonerModel() (*DeepseekReasonerModel, error) {
	model := NewDeepseekReasonerModel("")
	return &model, nil
}

func (f *Factory) createDeepseekChatModel() (*DeepseekChatModel, error) {
	model := NewDeepseekChatModel("")
	return &model, nil
}
