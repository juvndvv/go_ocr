package prompter

import "go_ocr/src/ocr/prompter/prompterTypes"

type Prompter interface {
	BuildPrompt(prompt string) (prompterTypes.Prompt, error)
	GetContext() string
}
