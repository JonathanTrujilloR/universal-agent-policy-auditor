package model

import (
	"strings"
	"testing"
)

func TestPermissionCopiesInputsAndRedactsEvidence(t *testing.T) {
	trace := NewResolutionTrace([]TraceStep{{Kind: TraceDefault, After: EffectAsk, Rules: []string{"default"}, Rationale: "documented default", Evidence: NewEvidence("docs", "permissions", "token=secret")}}, []string{"interactive approval"}, StatusIncomplete)
	permission := NewPermission("bash", EffectAsk, "project", "git *", []string{"trusted workspace"}, []Extension{{Target: "opencode", Key: "tool", Value: "bash"}}, trace)
	conditions := permission.Conditions()
	conditions[0] = "changed"
	steps := permission.Trace().Steps()
	steps[0].Rules[0] = "changed"
	if got := permission.Conditions()[0]; got != "trusted workspace" {
		t.Fatalf("conditions leaked mutation: %q", got)
	}
	if got := permission.Trace().Steps()[0].Rules[0]; got != "default" {
		t.Fatalf("trace leaked mutation: %q", got)
	}
	if got := permission.Trace().Steps()[0].Evidence.Value(); strings.Contains(got, "secret") {
		t.Fatalf("evidence leaked secret: %q", got)
	}
}

func TestResolutionTracePreservesOrderedStepsAndStatus(t *testing.T) {
	trace := NewResolutionTrace([]TraceStep{{Kind: TraceMerge, Before: EffectAllow, After: EffectAsk}, {Kind: TracePrecedence, Before: EffectAsk, After: EffectDeny}}, []string{"sandbox"}, StatusIncomplete)
	if got := trace.Steps(); len(got) != 2 || got[0].Kind != TraceMerge || got[1].After != EffectDeny {
		t.Fatalf("steps = %+v", got)
	}
	if trace.Status() != StatusIncomplete || trace.Unresolved()[0] != "sandbox" {
		t.Fatalf("status=%q unresolved=%v", trace.Status(), trace.Unresolved())
	}
}

func TestPermissionPreservesTypedProvenanceAndConditionalEffects(t *testing.T) {
	provenance := []Provenance{NewProvenance("claude-code", "settings.json", "secret")}
	permission := NewPermissionWithProvenance("Bash", EffectConditional, "managed", "Bash(git *)", []Condition{"workspace trust"}, provenance, nil, ResolutionTrace{})
	provenance[0] = NewProvenance("changed", "changed", "changed")
	if permission.Effect() != EffectConditional || permission.Conditions()[0] != Condition("workspace trust") {
		t.Fatalf("permission=%+v", permission)
	}
	if got := permission.Provenance()[0].Source(); got != "claude-code" {
		t.Fatalf("provenance leaked mutation: %q", got)
	}
}
