package calls

import (
    // "os"
    "fmt"
    "context"

    "github.com/openai/openai-go/v2"
    "github.com/openai/openai-go/v2/option"
    // "github.com/openai/openai-go/v2/shared"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
)

func Query4_1Nano(msg string) (string, error) {
    // Load API key
    client := openai.NewClient(
        option.WithAPIKey(config.OpenAIKey),
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

    return chatCompletion.Choices[0].Message.Content, nil
}

