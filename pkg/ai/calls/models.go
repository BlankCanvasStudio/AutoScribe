package calls

import (
	"os"
	"fmt"
	"context"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	// "github.com/openai/openai-go/v2/shared"

	log "github.com/sirupsen/logrus"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/types"
)

/*
*
 * Sends a message to the GPT-4.1 Nano model via OpenAI API and returns the response.
 *
 * Signature:
 * func Query4_1Nano(msg string) (string, error)
 *
 * Parameters:
 * - msg: string; the input message to send to the model.
 *
 * Returns:
 * - string: the content of the model's reply.
 * - error: non-nil if the API request fails.
 *
 * Errors/Exceptions:
 * - Returns an error if the API request to OpenAI fails.
 *
 * Side Effects:
 * - Creates an OpenAI client.
 * - Makes a network request to the OpenAI API.
 * - May append additional prompts from configuration.
 *
 * Edge Cases & Assumptions:
 * - Assumes the API key and configuration are properly set.
 * - Expects the API to return at least one choice.

*/
func Query4_1Nano(directive types.Directive, msg string) (string, error) {
	// Load API key
	client := openai.NewClient(
		option.WithAPIKey(directive.ApiKey),
	)

        /*
	if config.AdditionalPrompt != "" {
		msg += fmt.Sprintf("\n-----------------------\nAdditionally:\n%v\n", config.AdditionalPrompt)
	}
        */

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



/*
*
 * Reads a prompt template from a file, formats it with given arguments, and sends it to the specified AI model.
 *
 * Signature:
 * func QueryFromFile(model ai.Model, filename string, args ...any) (string, error)
 *
 * Parameters:
 * - model: ai.Model; the AI model to use.
 * - filename: string; path to the file containing the prompt template.
 * - args: ...any; arguments to format the prompt template.
 *
 * Returns:
 * - string: the response from the AI model.
 * - error: if reading the file or querying the model fails.
 *
 * Errors/Exceptions:
 * - Returns an error if reading the prompt file fails.
 * - Returns an error if the specified model does not exist.
 * - Returns an error if the API call to the model fails.
 *
 * Side Effects:
 * - Reads a file from disk.
 * - Logs the full prompt.
 * - Makes an API call to the AI service.
 *
 * Edge Cases & Assumptions:
 * - Assumes the prompt file exists and contains a valid format string.
 * - Assumes the model is supported.
 * - Assumes the API key and configuration are correctly set up.

*/
func QueryFromDirective(directive types.Directive, args ...any) (string, error) {
    param_prompt := directive.PromptText

    if directive.PromptText == "" {
        param_prompt_b, err := os.ReadFile(directive.Prompt)
        if err != nil {
            return "", fmt.Errorf("failed to read %v: %v", directive.Prompt, err)
        }

        param_prompt = string(param_prompt_b)
    }

    full_prompt := fmt.Sprintf(string(param_prompt), args...)

    log.Debugf("Full gpt prompt:\n%v\n", full_prompt)

    switch directive.Model {
        case types.GPT_41_Nano:
            return Query4_1Nano(directive, full_prompt)
        default:
            return "", fmt.Errorf("model %v doesn't exist", directive.Model)
    }
}

func QueryFromFile(directive types.Directive, args ...any) (string, error) {
    param_prompt, err := os.ReadFile(directive.Prompt)
    if err != nil {
        return "", fmt.Errorf("failed to read %v: %v", directive.Prompt, err)
    }

    full_prompt := fmt.Sprintf(string(param_prompt), args...)

    log.Debugf("Full gpt prompt:\n%v\n", full_prompt)

    switch directive.Model {
        case types.GPT_41_Nano:
            return Query4_1Nano(directive, full_prompt)
        default:
            return "", fmt.Errorf("model %v doesn't exist", directive.Model)
    }
}


/*
func QueryFromText(model types.Model, promptText string, args ...any) (string, error) {
    full_prompt := fmt.Sprintf(promptText, args...)

    log.Debugf("Full gpt prompt:\n%v\n", full_prompt)

    switch model {
        case types.GPT_41_Nano:
            return Query4_1Nano(full_prompt)
        default:
            return "", fmt.Errorf("model %v doesn't exist", model)
    }
}
*/

