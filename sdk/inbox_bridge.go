package sdk

import "encoding/json"

const inboxCreateSubject = "TW.XX.inbox.command.create"

func mapToInboxCreatePayload(payload any) (map[string]any, bool) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, false
	}

	var p struct {
		UserID      string `json:"userId"`
		MessageID   string `json:"messageId"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Category    string `json:"category"`
		Box         string `json:"box"`
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return nil, false
	}

	if p.UserID == "" || p.MessageID == "" || p.Title == "" || p.Description == "" || p.Category == "" || p.Box == "" {
		return nil, false
	}

	return map[string]any{
		"userId":      p.UserID,
		"messageId":   p.MessageID,
		"title":       p.Title,
		"description": p.Description,
		"category":    p.Category,
		"box":         p.Box,
	}, true
}
