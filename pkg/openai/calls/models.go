package calls

import (
	// "os"
	"context"
	"fmt"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	// "github.com/openai/openai-go/v2/shared"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/config"
)

/*
*
 * Executes a chat completion request to the OpenAI API using the GPT-4.1 Nano model.
 *
 * Summary:
 * Sends a message to the OpenAI GPT-4.1 Nano model, optionally appending an additional prompt,
 * and returns the generated response content.
 *
 * Signature:
 * func Query4_1Nano(msg string) (string, error)
 *
 * Parameters:
 * - msg (string): The user message to send to the model.
 *
 * Returns:
 * - string: The content of the first choice's message from the model's response.
 * - error: An error if the request fails.
 *
 * Errors/Exceptions:
 * - Returns an error if the API request fails.
 *
 * Side Effects:
 * - Creates an API client.
 * - Sends a request over the network.
 *
 * Edge Cases & Assumptions:
 * - Assumes `config.OpenAIKey` is a valid API key.
 * - Appends `config.AdditionalPrompt` if it's not empty.
 * - Uses context.TODO() for request context.

*/
func Query4_1Nano(msg string) (string, error) {
	// Load API key
	client := openai.NewClient(
		option.WithAPIKey(config.OpenAIKey),
	)

	if config.AdditionalPrompt != "" {
		msg += fmt.Sprintf("\n-----------------------\nAdditionally:\n%v\n", config.AdditionalPrompt)
	}

	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(msg),
		},
		Model: openai.ChatModelGPT4_1Nano,
	})

	if err != nil {
		return "", fmt.Errorf("failed to query 4.1 nano : %v", err)
	}

	return chatCompletion.Choices[0].Message.Content, nil
}
