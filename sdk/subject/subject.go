package subject

import (
	"fmt"
	"regexp"
	"strings"
)

type Type string

type Scope string

const (
	TypeCmd Type = "cmd"
	TypeEvt Type = "evt"
	TypeRun Type = "run"
)

const (
	ScopeApp   Scope = "app"
	ScopeSpace Scope = "space"
	ScopeUser  Scope = "user"
)

type Subject struct {
	Type     Type
	Scope    Scope
	ScopeID  string
	Resource string
	Action   string
}

var (
	tokenPattern      = regexp.MustCompile(`^[a-z0-9]+(?:[a-z0-9_-]*[a-z0-9])?$`)
	spaceScopePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{10}$`)
)

func Build(s Subject) (string, error) {
	if err := validateSubject(s); err != nil {
		return "", err
	}
	return strings.Join([]string{string(s.Type), string(s.Scope), s.ScopeID, s.Resource, s.Action}, "."), nil
}

func MustBuild(s Subject) string {
	built, err := Build(s)
	if err != nil {
		panic(err)
	}
	return built
}

func Validate(value string) error {
	_, err := Parse(value)
	return err
}

func Parse(value string) (Subject, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 5 {
		return Subject{}, fmt.Errorf("invalid subject format: expected 5 segments, got %d", len(parts))
	}

	s := Subject{
		Type:     Type(parts[0]),
		Scope:    Scope(parts[1]),
		ScopeID:  parts[2],
		Resource: parts[3],
		Action:   parts[4],
	}
	if err := validateSubject(s); err != nil {
		return Subject{}, err
	}
	return s, nil
}

func CmdApp(app, resource, action string) (string, error) {
	return Build(Subject{Type: TypeCmd, Scope: ScopeApp, ScopeID: app, Resource: resource, Action: action})
}

func CmdSpace(spaceID, resource, action string) (string, error) {
	return Build(Subject{Type: TypeCmd, Scope: ScopeSpace, ScopeID: spaceID, Resource: resource, Action: action})
}

func CmdUser(account, resource, action string) (string, error) {
	return Build(Subject{Type: TypeCmd, Scope: ScopeUser, ScopeID: account, Resource: resource, Action: action})
}

func EvtApp(app, resource, action string) (string, error) {
	return Build(Subject{Type: TypeEvt, Scope: ScopeApp, ScopeID: app, Resource: resource, Action: action})
}

func EvtSpace(spaceID, resource, action string) (string, error) {
	return Build(Subject{Type: TypeEvt, Scope: ScopeSpace, ScopeID: spaceID, Resource: resource, Action: action})
}

func EvtUser(account, resource, action string) (string, error) {
	return Build(Subject{Type: TypeEvt, Scope: ScopeUser, ScopeID: account, Resource: resource, Action: action})
}

func RunApp(app, resource, action string) (string, error) {
	return Build(Subject{Type: TypeRun, Scope: ScopeApp, ScopeID: app, Resource: resource, Action: action})
}

func RunSpace(spaceID, resource, action string) (string, error) {
	return Build(Subject{Type: TypeRun, Scope: ScopeSpace, ScopeID: spaceID, Resource: resource, Action: action})
}

func RunUser(account, resource, action string) (string, error) {
	return Build(Subject{Type: TypeRun, Scope: ScopeUser, ScopeID: account, Resource: resource, Action: action})
}

func validateSubject(s Subject) error {
	if !isValidType(s.Type) {
		return fmt.Errorf("invalid type: %s", s.Type)
	}
	if !isValidScope(s.Scope) {
		return fmt.Errorf("invalid scope: %s", s.Scope)
	}
	if err := validateScopeID(s.Scope, s.ScopeID); err != nil {
		return err
	}
	if !isToken(s.Resource) {
		return fmt.Errorf("invalid resource: %s", s.Resource)
	}
	if !isToken(s.Action) {
		return fmt.Errorf("invalid action: %s", s.Action)
	}
	return nil
}

func validateScopeID(scope Scope, scopeID string) error {
	if scope == ScopeSpace {
		if !spaceScopePattern.MatchString(scopeID) {
			return fmt.Errorf("invalid scope_id for space: %s", scopeID)
		}
		return nil
	}
	if !isToken(scopeID) {
		return fmt.Errorf("invalid scope_id: %s", scopeID)
	}
	return nil
}

func isValidType(t Type) bool {
	switch t {
	case TypeCmd, TypeEvt, TypeRun:
		return true
	default:
		return false
	}
}

func isValidScope(scope Scope) bool {
	switch scope {
	case ScopeApp, ScopeSpace, ScopeUser:
		return true
	default:
		return false
	}
}

func isToken(value string) bool {
	return tokenPattern.MatchString(value)
}
