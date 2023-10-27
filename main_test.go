package main

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func createLocalClient() *openai.Client {
	config := openai.DefaultConfig("test")
	config.BaseURL = "http://localhost:9200/v1"
	client := openai.NewClientWithConfig(config)
	return client
}

func Test_ProxyServer(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("OPENAI_API_KEY is not set")
	}
	go main()
	c := createLocalClient()

	t.Run("Basic", func(t *testing.T) {
		t.Parallel()
		resp, err := c.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: openai.GPT3Dot5Turbo,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: `Say "Hello, World!"`,
					},
				},
			},
		)
		assert.Nil(t, err)
		assert.Equal(t, "Hello, World!", resp.Choices[0].Message.Content)
	})

	t.Run("Stream", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		req := openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo,
			MaxTokens: 20,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: `Say "Hello, World!"`,
				},
			},
			Stream: true,
		}
		stream, err := c.CreateChatCompletionStream(ctx, req)
		assert.Nil(t, err)
		defer stream.Close()

		var finalResponse strings.Builder
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			assert.Nil(t, err)
			finalResponse.WriteString(response.Choices[0].Delta.Content)
		}
		assert.Equal(t, "Hello, World!", finalResponse.String())
	})
}
