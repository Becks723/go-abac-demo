package abac

import (
	"fmt"
	"sync"
)

type Enforcer struct {
	mu    sync.RWMutex
	rules []Rule
}

func NewEnforcer(rules ...Rule) *Enforcer {
	enforcer := &Enforcer{}
	for _, rule := range rules {
		_ = enforcer.AddRule(rule)
	}
	return enforcer
}

func (e *Enforcer) Enforce(ctx AccessContext) Decision {
	rules := e.snapshotRules()
	return evaluateRules(ctx, rules)
}

func (e *Enforcer) AddRule(rule Rule) error {
	if rule == nil {
		return fmt.Errorf("rule is nil")
	}
	if rule.ID() == "" {
		return fmt.Errorf("rule id is empty")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	for _, existing := range e.rules {
		if existing.ID() == rule.ID() {
			return fmt.Errorf("rule already exists: %s", rule.ID())
		}
	}

	e.rules = append(e.rules, rule)
	return nil
}

func (e *Enforcer) RemoveRule(id string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	for i, rule := range e.rules {
		if rule.ID() == id {
			e.rules = append(e.rules[:i], e.rules[i+1:]...)
			return true
		}
	}
	return false
}

func (e *Enforcer) RuleIDs() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ids := make([]string, 0, len(e.rules))
	for _, rule := range e.rules {
		ids = append(ids, rule.ID())
	}
	return ids
}

func (e *Enforcer) snapshotRules() []Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()

	rules := make([]Rule, len(e.rules))
	copy(rules, e.rules)
	return rules
}
