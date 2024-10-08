package gpt

import (
	"errors"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksgpt "github.com/lmquang/code-review/mocks/pkg/gpt"
	mocksgptopenai "github.com/lmquang/code-review/mocks/pkg/gpt/openai"
)

func TestNewOpenAIClient(t *testing.T) {
	client := NewOpenAIClient("test-api-key")

	assert.NotNil(t, client)
	assert.Implements(t, (*IGPT)(nil), client)
}

func TestGPT_Review(t *testing.T) {
	tests := []struct {
		name            string
		originalContent string
		formattedDiff   string
		mockResponse    openai.ChatCompletionResponse
		mockError       error
		expectedError   error
	}{
		{
			name:            "Successful review",
			originalContent: "<original-content><file path=\"file.txt\">Line 1\nLine 2\nLine 3</file></original-content>",
			formattedDiff:   "<git-diff><file><n>file.txt</n><changes><![CDATA[diff --git a/file.txt b/file.txt\nindex 1234567..890abcd 100644\n--- a/file.txt\n+++ b/file.txt\n@@ -1,3 +1,4 @@\n Line 1\n-Line 2\n+Updated Line 2\n Line 3\n+New Line 4]]></changes></file></git-diff>",
			mockResponse: openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "<review><style_and_conventions>Style is consistent.</style_and_conventions><comments_review>No comments added.</comments_review><best_practices>Code follows best practices.</best_practices><summary>Changes look good.</summary></review>",
						},
					},
				},
			},
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:            "Error during review",
			originalContent: "<original-content><file path=\"file.txt\">Original content</file></original-content>",
			formattedDiff:   "<git-diff><file><n>file.txt</n><changes><![CDATA[Sample diff]]></changes></file></git-diff>",
			mockResponse:    openai.ChatCompletionResponse{},
			mockError:       errors.New("OpenAI API error"),
			expectedError:   errors.New("ChatCompletion error: OpenAI API error"),
		},
		{
			name:            "Empty formatted diff",
			originalContent: "<original-content></original-content>",
			formattedDiff:   "<git-diff></git-diff>",
			mockResponse: openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "<review><summary>No changes detected in the diff.</summary></review>",
						},
					},
				},
			},
			mockError:     nil,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOpenAI := new(mocksgptopenai.IOpenAI)
			mockOpenAI.On("CreateChatCompletion", mock.Anything, mock.MatchedBy(func(req openai.ChatCompletionRequest) bool {
				return req.MaxTokens == 1000 && req.Model == openai.GPT4oMini &&
					strings.Contains(req.Messages[0].Content, tt.originalContent) &&
					req.Messages[1].Content == tt.formattedDiff
			})).Return(tt.mockResponse, tt.mockError)
			mockOpenAI.On("GetModel").Return(openai.GPT4oMini)

			gpt := &gpt{
				client: mockOpenAI,
			}

			result, err := gpt.Review(tt.originalContent, tt.formattedDiff)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.Choices[0].Message.Content, result)
			}

			mockOpenAI.AssertExpectations(t)
		})
	}
}

func TestGPT_Client(t *testing.T) {
	mockOpenAI := new(mocksgptopenai.IOpenAI)
	gpt := &gpt{
		client: mockOpenAI,
	}

	assert.Equal(t, mockOpenAI, gpt.Client())
}

func TestGPT_ImplementsIGPT(t *testing.T) {
	client := NewOpenAIClient("test-api-key")
	assert.Implements(t, (*IGPT)(nil), client)
}

func TestMockIGPT(t *testing.T) {
	mockGPT := new(mocksgpt.IGPT)
	mockOpenAI := new(mocksgptopenai.IOpenAI)

	originalContent := "<original-content><file path=\"file.txt\">Original content</file></original-content>"
	formattedDiff := "<git-diff><file><n>file.txt</n><changes><![CDATA[Sample diff]]></changes></file></git-diff>"

	mockGPT.On("Review", originalContent, formattedDiff).Return("<review><summary>Mock review</summary></review>", nil)
	mockGPT.On("Client").Return(mockOpenAI)

	result, err := mockGPT.Review(originalContent, formattedDiff)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(result, "Mock review"))

	client := mockGPT.Client()
	assert.Equal(t, mockOpenAI, client)

	mockGPT.AssertExpectations(t)
}

func TestOpenAIClient_GetModel(t *testing.T) {
	mockOpenAI := new(mocksgptopenai.IOpenAI)
	mockOpenAI.On("GetModel").Return(openai.GPT4oMini)

	gpt := &gpt{
		client: mockOpenAI,
	}

	model := gpt.Client().GetModel()
	assert.Equal(t, openai.GPT4oMini, model)

	mockOpenAI.AssertExpectations(t)
}
