package call;

import (
    "fmt"
    "context"
    "slices"

    "github.com/openai/openai-go/v2"
    "github.com/openai/openai-go/v2/option"

)

func Query41Nano(msg string, ApiKey string, validOutputs []string) (string, error) {

    res := ""

    for !slices.Contains(validOutputs, res) {
        // Load API key
        client := openai.NewClient(
                option.WithAPIKey(ApiKey),
        )

        chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
                Messages: []openai.ChatCompletionMessageParamUnion{
                        openai.UserMessage(msg),
                },
                Model: openai.ChatModelGPT4_1Nano,
        })

        if err != nil {
                return "", fmt.Errorf("failed to query 4.1 nano : %v", err)
        }

        res = chatCompletion.Choices[0].Message.Content
    }

    return res, nil
}

