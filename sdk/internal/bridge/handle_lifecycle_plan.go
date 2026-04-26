package bridge

import (
	"fmt"

	"github.com/include2md/eventsdk/sdk/subject"
)

func BuildHandleLifecycleRules(publisher Publisher, handleSubject string) ([]Rule, error) {
	parsed, err := subject.Parse(handleSubject)
	if err != nil {
		return nil, fmt.Errorf("parse handle subject: %w", err)
	}
	if parsed.Scope != subject.ScopeApp {
		return nil, fmt.Errorf("handle subject must use app scope")
	}

	app := parsed.ScopeID
	beforeSubject, err := subject.RunApp(app, "execution", "started")
	if err != nil {
		return nil, fmt.Errorf("build before subject: %w", err)
	}
	afterSubject, err := subject.EvtApp(app, parsed.Resource, parsed.Action)
	if err != nil {
		return nil, fmt.Errorf("build after subject: %w", err)
	}

	beforeRule := NewHandlePublishRule(publisher, HandlePublishRuleOptions{
		Name:     "handle-execution-started",
		Stage:    StageBeforeHandle,
		Priority: 10,
		Subject:  beforeSubject,
	})
	afterRule := NewHandlePublishRule(publisher, HandlePublishRuleOptions{
		Name:     "handle-execution-finished",
		Stage:    StageAfterHandle,
		Priority: 10,
		Subject:  afterSubject,
	})
	return []Rule{beforeRule, afterRule}, nil
}
