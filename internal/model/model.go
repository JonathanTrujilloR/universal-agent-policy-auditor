// Package model defines immutable facts used to report native policy semantics.
package model

import (
	"crypto/sha256"
	"encoding/hex"
)

type Effect string

const (
	EffectAllow       Effect = "allow"
	EffectDeny        Effect = "deny"
	EffectAsk         Effect = "ask"
	EffectInherited   Effect = "inherited"
	EffectConditional Effect = "conditional"
	EffectUnresolved  Effect = "unresolved"
)

type Scope string
type Condition string
type Completeness string

const (
	CompletenessComplete   Completeness = "complete"
	CompletenessIncomplete Completeness = "incomplete"
	StatusIncomplete                    = CompletenessIncomplete
)

type Evidence struct{ source, location, value string }

func NewEvidence(source, location, raw string) Evidence {
	digest := sha256.Sum256([]byte(raw))
	return Evidence{source: source, location: location, value: "sha256:" + hex.EncodeToString(digest[:])}
}
func (e Evidence) Source() string   { return e.source }
func (e Evidence) Location() string { return e.location }
func (e Evidence) Value() string    { return e.value }

type Provenance = Evidence

func NewProvenance(source, location, raw string) Provenance {
	return NewEvidence(source, location, raw)
}

type TraceKind string

const (
	TraceRule       TraceKind = "rule"
	TraceDefault    TraceKind = "default"
	TraceMerge      TraceKind = "merge"
	TracePrecedence TraceKind = "precedence"
	TraceRuntime    TraceKind = "runtime-dependency"
)

type TraceStep struct {
	Kind      TraceKind
	Before    Effect
	After     Effect
	Rules     []string
	Rationale string
	Evidence  Evidence
}

type ResolutionTrace struct {
	steps      []TraceStep
	unresolved []string
	status     Completeness
}

func NewResolutionTrace(steps []TraceStep, unresolved []string, status Completeness) ResolutionTrace {
	return ResolutionTrace{steps: copySteps(steps), unresolved: append([]string(nil), unresolved...), status: status}
}
func (t ResolutionTrace) Steps() []TraceStep   { return copySteps(t.steps) }
func (t ResolutionTrace) Unresolved() []string { return append([]string(nil), t.unresolved...) }
func (t ResolutionTrace) Status() Completeness { return t.status }
func copySteps(in []TraceStep) []TraceStep {
	out := append([]TraceStep(nil), in...)
	for i := range out {
		out[i].Rules = append([]string(nil), in[i].Rules...)
	}
	return out
}

type Extension struct{ Target, Key, Value string }
type Permission struct {
	capability string
	effect     Effect
	scope      Scope
	matcher    string
	conditions []Condition
	extensions []Extension
	provenance []Provenance
	trace      ResolutionTrace
}

func NewPermission(capability string, effect Effect, scope Scope, matcher string, conditions []string, extensions []Extension, trace ResolutionTrace) Permission {
	typed := make([]Condition, len(conditions))
	for i := range conditions {
		typed[i] = Condition(conditions[i])
	}
	return NewPermissionWithProvenance(capability, effect, scope, matcher, typed, nil, extensions, trace)
}
func NewPermissionWithProvenance(capability string, effect Effect, scope Scope, matcher string, conditions []Condition, provenance []Provenance, extensions []Extension, trace ResolutionTrace) Permission {
	return Permission{capability: capability, effect: effect, scope: scope, matcher: matcher, conditions: append([]Condition(nil), conditions...), provenance: append([]Provenance(nil), provenance...), extensions: append([]Extension(nil), extensions...), trace: NewResolutionTrace(trace.steps, trace.unresolved, trace.status)}
}
func (p Permission) Capability() string       { return p.capability }
func (p Permission) Effect() Effect           { return p.effect }
func (p Permission) Scope() Scope             { return p.scope }
func (p Permission) Matcher() string          { return p.matcher }
func (p Permission) Conditions() []Condition  { return append([]Condition(nil), p.conditions...) }
func (p Permission) Extensions() []Extension  { return append([]Extension(nil), p.extensions...) }
func (p Permission) Provenance() []Provenance { return append([]Provenance(nil), p.provenance...) }
func (p Permission) Trace() ResolutionTrace {
	return NewResolutionTrace(p.trace.steps, p.trace.unresolved, p.trace.status)
}

type Finding struct{ code, message string }

func NewFinding(code, message string) Finding { return Finding{code: code, message: message} }
func (f Finding) Code() string                { return f.code }
func (f Finding) Message() string             { return f.message }
