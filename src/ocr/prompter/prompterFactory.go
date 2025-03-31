package prompter

import (
	"fmt"
	"go_ocr/src/ocr/prompter/prompterImplementations"
)

type PrompterFactory struct {
}

func NewPrompterFactory() *PrompterFactory {
	return &PrompterFactory{}
}

func (pf *PrompterFactory) GetPrompter(prompterType string) (Prompter, error) {
	switch prompterType {
	case "payroll":
		return prompterImplementations.NewPayrollPrompter(), nil
	default:
		return nil, fmt.Errorf("prompter type not found")
	}
}
