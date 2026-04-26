package subject_test

import (
	"testing"

	"github.com/include2md/eventsdk/sdk/subject"
)

func TestBuild(t *testing.T) {
	s, err := subject.Build(subject.Subject{
		Type:     subject.TypeCmd,
		Scope:    subject.ScopeApp,
		ScopeID:  "tdemo",
		Resource: "todo",
		Action:   "create",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if s != "cmd.app.tdemo.todo.create" {
		t.Fatalf("unexpected subject: %s", s)
	}
}

func TestBuildRejectsInvalidScopeIDForSpace(t *testing.T) {
	_, err := subject.Build(subject.Subject{
		Type:     subject.TypeEvt,
		Scope:    subject.ScopeSpace,
		ScopeID:  "8f3ab2d1",
		Resource: "doc",
		Action:   "updated",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidate(t *testing.T) {
	if err := subject.Validate("evt.user.dennissu.profile.updated"); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestValidateRejectsInvalidFormat(t *testing.T) {
	if err := subject.Validate("evt.user.dennissu.profile"); err == nil {
		t.Fatal("expected error")
	}
}

func TestParse(t *testing.T) {
	s, err := subject.Parse("run.app.tdemo.execution.started")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if s.Type != subject.TypeRun || s.Scope != subject.ScopeApp || s.ScopeID != "tdemo" || s.Resource != "execution" || s.Action != "started" {
		t.Fatalf("unexpected parsed subject: %+v", s)
	}
}

func TestHelpers(t *testing.T) {
	cmd, err := subject.CmdApp("tdemo", "todo", "list")
	if err != nil {
		t.Fatalf("cmd app err: %v", err)
	}
	if cmd != "cmd.app.tdemo.todo.list" {
		t.Fatalf("unexpected cmd: %s", cmd)
	}

	evt, err := subject.EvtUser("dennissu", "profile", "updated")
	if err != nil {
		t.Fatalf("evt user err: %v", err)
	}
	if evt != "evt.user.dennissu.profile.updated" {
		t.Fatalf("unexpected evt: %s", evt)
	}

	run, err := subject.RunApp("tdemo", "execution", "started")
	if err != nil {
		t.Fatalf("run app err: %v", err)
	}
	if run != "run.app.tdemo.execution.started" {
		t.Fatalf("unexpected run: %s", run)
	}
}
