package gpt

import (
	"context"
	"fmt"
	"log"

	openai "github.com/sashabaranov/go-openai"
)

// Client represents a GPT client
type Client struct {
	client *openai.Client
	model  string
}

// NewClient creates a new GPT client
func NewClient(apiKey string) *Client {
	client := openai.NewClient(apiKey)
	return &Client{
		client: client,
		model:  openai.GPT4oMini,
	}
}

// Review sends the formatted diff to GPT for review
func (c *Client) Review(formattedDiff string) (string, error) {
	prompt := `You are an AI assistant tasked with reviewing code changes based on a git diff output. Your goal is to ensure the code follows the existing style and conventions of the codebase, while also suggesting improvements to align with best practices. Follow these instructions to complete the review:

1. First, you will be provided with the git diff output in XML format: <git_diff>{{CODE_DIFF}}</git_diff>

2. Detect language in the code changes

3. Review the code changes for style and conventions:
   a. Analyze the existing code style in the diff output.
   b. Check if the new changes follow the same style and conventions.
   c. Look for inconsistencies in indentation, naming conventions, and code structure.

4. Check for comments in the changes:
   a. Identify any new or modified comments.
   b. Evaluate if the comments are clear, concise, and provide valuable information.
   c. Check if comments are up-to-date with the code changes.

5. Suggest improvements based on best practices:
   a. Identify any code patterns or practices that could be improved.
   b. Recommend changes that align with the best practices for the specified programming language.
   c. Provide explanations for why these changes would be beneficial.

6. Provide your review in the following format:
   <review>
   <style_and_conventions>
   [List observations about code style and conventions, including any inconsistencies or areas for improvement]
   </style_and_conventions>

   <comments_review>
   [Provide feedback on the comments in the code changes]
   </comments_review>

   <best_practices>
   [Suggest improvements based on best practices, explaining the benefits of each suggestion]
   </best_practices>

   <summary>
   [Provide a brief summary of the overall code changes and your main recommendations]
   </summary>

   <suggest_changes>
	[List files and lines where changes are suggested, along with the recommended modifications]
   </suggest_changes>
   </review>

Remember to be constructive in your feedback and provide clear explanations for your suggestions. Focus on maintaining consistency with the existing codebase while promoting best practices for the specified programming language.`

	log.Printf("Sending to GPT (%v)\n", c.model)
	resp, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: prompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: formattedDiff,
				},
			},
			MaxTokens: 1000,
		},
	)

	if err != nil {
		return "", fmt.Errorf("ChatCompletion error: %v", err)
	}

	return resp.Choices[0].Message.Content, nil
}

func (c *Client) SetModel(model string) {
	c.model = model
}
