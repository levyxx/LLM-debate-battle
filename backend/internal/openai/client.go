package openai

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Message struct {
	Role    string
	Content string
}

type Client struct {
	client *openai.Client
	model  string
}

func NewClient(apiKey, model string) *Client {
	apiKey = strings.TrimSpace(apiKey)
	model = strings.TrimSpace(model)

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &Client{
		client: &client,
		model:  model,
	}
}

// 構造化出力を使用したチャット補完
func (c *Client) ChatCompletionWithSchema(ctx context.Context, messages []Message, schemaName string, schema map[string]any) (string, error) {
	if c.model == "" {
		err := errors.New("openai model is empty")
		log.Printf("[OpenAI] %v", err)
		return "", err
	}
	if len(messages) == 0 {
		err := errors.New("openai messages are empty")
		log.Printf("[OpenAI] %v", err)
		return "", err
	}

	started := time.Now()
	log.Printf("[OpenAI] ChatCompletion start model=%s messages=%d", c.model, len(messages))

	chatMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))
	for _, msg := range messages {
		switch strings.ToLower(msg.Role) {
		case "user":
			chatMessages = append(chatMessages, openai.UserMessage(msg.Content))
		case "assistant":
			chatMessages = append(chatMessages, openai.AssistantMessage(msg.Content))
		case "system":
			chatMessages = append(chatMessages, openai.SystemMessage(msg.Content))
		default:
			chatMessages = append(chatMessages, openai.UserMessage(msg.Content))
		}
	}

	completion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(c.model),
		Messages: chatMessages,
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				Type: "json_schema",
				JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:   schemaName,
					Schema: schema,
					Strict: openai.Bool(true),
				},
			},
		},
	})
	if err != nil {
		log.Printf("[OpenAI] request failed: %v", err)
		return "", err
	}

	if len(completion.Choices) == 0 {
		err := errors.New("openai response had no choices")
		log.Printf("[OpenAI] %v", err)
		return "", err
	}

	log.Printf("[OpenAI] ChatCompletion success duration=%s", time.Since(started))
	return completion.Choices[0].Message.Content, nil
}

// 通常のチャット補完（構造化出力なし）
func (c *Client) ChatCompletion(ctx context.Context, messages []Message) (string, error) {
	if c.model == "" {
		err := errors.New("openai model is empty")
		log.Printf("[OpenAI] %v", err)
		return "", err
	}
	if len(messages) == 0 {
		err := errors.New("openai messages are empty")
		log.Printf("[OpenAI] %v", err)
		return "", err
	}

	started := time.Now()
	log.Printf("[OpenAI] ChatCompletion start model=%s messages=%d", c.model, len(messages))

	chatMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))
	for _, msg := range messages {
		switch strings.ToLower(msg.Role) {
		case "user":
			chatMessages = append(chatMessages, openai.UserMessage(msg.Content))
		case "assistant":
			chatMessages = append(chatMessages, openai.AssistantMessage(msg.Content))
		case "system":
			chatMessages = append(chatMessages, openai.SystemMessage(msg.Content))
		default:
			chatMessages = append(chatMessages, openai.UserMessage(msg.Content))
		}
	}

	completion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(c.model),
		Messages: chatMessages,
	})
	if err != nil {
		log.Printf("[OpenAI] request failed: %v", err)
		return "", err
	}

	if len(completion.Choices) == 0 {
		err := errors.New("openai response had no choices")
		log.Printf("[OpenAI] %v", err)
		return "", err
	}

	log.Printf("[OpenAI] ChatCompletion success duration=%s", time.Since(started))
	return completion.Choices[0].Message.Content, nil
}
