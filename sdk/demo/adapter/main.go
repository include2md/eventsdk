package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/include2md/eventsdk/sdk/bootstrap"
)

type commandRequest struct {
	UserID      string `json:"userId"`
	MessageID   string `json:"messageId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Box         string `json:"box"`
}

type adapterReply struct {
	OK       bool            `json:"ok"`
	Source   string          `json:"source"`
	UserID   string          `json:"userId"`
	RestData json.RawMessage `json:"restData"`
}

func main() {
	ctx := context.Background()
	natsURL := envOr("NATS_URL", "nats://127.0.0.1:4222")
	namespace := envOr("SDK_NAMESPACE", "TW.XX")
	mockBaseURL := envOr("MOCK_API_BASE_URL", "http://127.0.0.1:18080")
	commandSubject := fmt.Sprintf("%s.user.command.create", namespace)

	service, err := bootstrap.NewService(bootstrap.Options{NATSURL: natsURL})
	if err != nil {
		log.Fatalf("new service: %v", err)
	}
	defer service.Close()
	log.Printf("adapter listening subject=%s nats=%s mock_api=%s", commandSubject, natsURL, mockBaseURL)

	err = service.HandleRequest(ctx, commandSubject, func(ctx context.Context, request []byte) ([]byte, error) {
		if len(request) == 0 {
			return nil, fmt.Errorf("empty request payload")
		}

		var req commandRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, fmt.Errorf("invalid request json: %w", err)
		}
		if err := validateCommandRequest(req); err != nil {
			return nil, err
		}

		httpClient := &http.Client{Timeout: 3 * time.Second}
		resp, err := httpClient.Get(fmt.Sprintf("%s/profile?id=%s", mockBaseURL, req.UserID))
		if err != nil {
			return nil, fmt.Errorf("call mock api: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read mock api body: %w", err)
		}

		reply, err := json.Marshal(adapterReply{
			OK:       true,
			Source:   "adapter",
			UserID:   req.UserID,
			RestData: body,
		})
		if err != nil {
			return nil, fmt.Errorf("marshal reply: %w", err)
		}

		log.Printf("processed request userId=%s messageId=%s", req.UserID, req.MessageID)
		return reply, nil
	})
	if err != nil {
		log.Fatalf("handle request: %v", err)
	}

	select {}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func validateCommandRequest(req commandRequest) error {
	switch {
	case req.UserID == "":
		return fmt.Errorf("missing field: userId")
	case req.MessageID == "":
		return fmt.Errorf("missing field: messageId")
	case req.Title == "":
		return fmt.Errorf("missing field: title")
	case req.Description == "":
		return fmt.Errorf("missing field: description")
	case req.Category == "":
		return fmt.Errorf("missing field: category")
	case req.Box == "":
		return fmt.Errorf("missing field: box")
	}
	return nil
}
