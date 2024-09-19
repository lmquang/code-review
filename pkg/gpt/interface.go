package gpt

import (
	gptopenai "github.com/lmquang/code-review/pkg/gpt/openai"
)

type IGPT interface {
	Review(formattedDiff string) (string, error)
	Client() gptopenai.IOpenAI
}

type gpt struct {
	client gptopenai.IOpenAI
}
