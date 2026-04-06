package subject

import "fmt"

type Resolver interface {
	CommandSubject(commandName string) (string, error)
	EventSubject(eventType string) string
	EventConsumeSubject() string
}

type resolver struct {
	namespace string
}

func NewResolver(namespace string) Resolver {
	return &resolver{namespace: namespace}
}

func (r *resolver) CommandSubject(commandName string) (string, error) {
	switch commandName {
	case "ListMessages":
		return fmt.Sprintf("%s.inbox.command.list", r.namespace), nil
	case "GetMessage":
		return fmt.Sprintf("%s.inbox.command.get", r.namespace), nil
	case "CreateMessage":
		return fmt.Sprintf("%s.inbox.command.create", r.namespace), nil
	case "ReadMessage":
		return fmt.Sprintf("%s.inbox.command.read", r.namespace), nil
	case "DeleteMessage":
		return fmt.Sprintf("%s.inbox.command.delete", r.namespace), nil
	default:
		return "", fmt.Errorf("unknown command: %s", commandName)
	}
}

func (r *resolver) EventSubject(eventType string) string {
	return fmt.Sprintf("%s.sdk.event.%s", r.namespace, eventType)
}

func (r *resolver) EventConsumeSubject() string {
	return fmt.Sprintf("%s.sdk.event.*", r.namespace)
}
