package gpt

import (
	"context"
	"fmt"
	"log"

	"github.com/sashabaranov/go-openai"

	gptopenai "github.com/lmquang/code-review/pkg/gpt/openai"
)

// NewOpenAIClient creates a new GPT client
func NewOpenAIClient(apiKey string) IGPT {
	return &gpt{
		client: gptopenai.NewOpenAI(openai.NewClient(apiKey), openai.GPT4oMini),
	}
}

func (c *gpt) Client() gptopenai.IOpenAI {
	return c.client
}

// Review sends the original content and formatted diff to GPT for review
func (c *gpt) Review(originalContent, formattedDiff string) (string, error) {
	prompt := fmt.Sprintf(`You are an AI assistant tasked with reviewing code changes based on the original content and a git diff output. Your goal is to ensure the code follows the existing style and conventions of the codebase, while also suggesting improvements to align with best practices. Follow these instructions to complete the review:

1. You will be provided with two pieces of information:
   a. The original content of the files before changes: <original-content>%v</original-content>
   b. The git diff output in XML format: <git-diff>{{CODE_DIFF}}</git-diff>

2. Analyze both the original content and the changes to:
   a. Understand the context of the changes
   b. Detect the programming language(s) used
   c. Identify the overall purpose and structure of the code

3. Review the code changes for style and conventions:
   a. Analyze the existing code style in the original content and diff output
   b. Check if the new changes follow the same style and conventions
   c. Look for inconsistencies in indentation, naming conventions, and code structure

4. Check for comments in the changes:
   a. Identify any new or modified comments
   b. Evaluate if the comments are clear, concise, and provide valuable information
   c. Check if comments are up-to-date with the code changes

5. Suggest improvements based on best practices:
   a. Identify any code patterns or practices that could be improved
   b. Recommend changes that align with the best practices for the specified programming language(s)
   c. Provide explanations for why these changes would be beneficial

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
   [List files and lines where changes are suggested, along with the recommended modifications, make sure file is not duplicated for each recommendation] as per the following example:
   <file>
     <n>file_name</n>
     <line>line_number</line>
     <change>proposed_change</change>
   </file>
   </suggest_changes>
   </review>

Remember to be constructive in your feedback and provide clear explanations for your suggestions. Focus on maintaining consistency with the existing codebase while promoting best practices for the specified programming language(s).`, originalContent)

	log.Printf("Sending %v characters to GPT (%v)\n", len(formattedDiff), c.client.GetModel())
	resp, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: c.client.GetModel(),
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
