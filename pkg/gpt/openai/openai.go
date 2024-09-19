package openai

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

// Client represents a GPT client
type openAI struct {
	client *openai.Client
	model  string
}

type IOpenAI interface {
	SetModel(model string)
	GetModel() string
	CreateChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

func NewOpenAI(client *openai.Client, model string) IOpenAI {
	return &openAI{
		client: client,
		model:  model,
	}
}

func (c *openAI) SetModel(model string) {
	c.model = model
}

func (c *openAI) GetModel() string {
	return c.model
}

func (c *openAI) CreateChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return c.client.CreateChatCompletion(ctx, request)
}
