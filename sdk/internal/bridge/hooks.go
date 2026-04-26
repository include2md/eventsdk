package bridge

import (
	"context"
	"errors"
	"sort"
)

type Stage string

const (
	StageBeforeHandle Stage = "before_handle"
	StageAfterHandle  Stage = "after_handle"
)

type PolicyMode string

const (
	PolicyFail   PolicyMode = "fail"
	PolicyIgnore PolicyMode = "ignore"
)

type Context struct {
	Stage    Stage
	Subject  string
	Request  []byte
	Response []byte
	Err      error
}

type Rule interface {
	Name() string
	Stage() Stage
	Priority() int
	Match(*Context) bool
	Run(context.Context, *Context) error
	Policy() *PolicyMode
}

type Options struct {
	DefaultPolicy PolicyMode
	Rules         []Rule
}

type Hooks struct {
	defaultPolicy PolicyMode
	rules         []Rule
}

func NewHooks(opts Options) *Hooks {
	defaultPolicy := opts.DefaultPolicy
	if defaultPolicy == "" {
		defaultPolicy = PolicyFail
	}

	rules := append([]Rule(nil), opts.Rules...)
	return &Hooks{
		defaultPolicy: defaultPolicy,
		rules:         rules,
	}
}

func (h *Hooks) Apply(ctx context.Context, bc *Context) error {
	if h == nil || bc == nil {
		return nil
	}

	matched := make([]Rule, 0, len(h.rules))
	for _, rule := range h.rules {
		if rule == nil || rule.Stage() != bc.Stage {
			continue
		}
		if !rule.Match(bc) {
			continue
		}
		matched = append(matched, rule)
	}

	sort.SliceStable(matched, func(i, j int) bool {
		return matched[i].Priority() < matched[j].Priority()
	})

	for _, rule := range matched {
		err := rule.Run(ctx, bc)
		if err == nil {
			continue
		}

		switch h.policyFor(rule) {
		case PolicyIgnore:
			continue
		case PolicyFail:
			return err
		default:
			return errors.New("unknown bridge policy")
		}
	}
	return nil
}

func (h *Hooks) WithExtraRules(extra []Rule) *Hooks {
	if h == nil {
		return NewHooks(Options{Rules: extra})
	}

	combined := append([]Rule(nil), h.rules...)
	combined = append(combined, extra...)
	return &Hooks{
		defaultPolicy: h.defaultPolicy,
		rules:         combined,
	}
}

func (h *Hooks) policyFor(rule Rule) PolicyMode {
	if rule == nil {
		return h.defaultPolicy
	}
	if p := rule.Policy(); p != nil && *p != "" {
		return *p
	}
	return h.defaultPolicy
}
