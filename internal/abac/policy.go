package abac

type Rule interface {
	ID() string
	Evaluate(ctx AccessContext) RuleResult
}

type RuleFunc struct {
	id string
	fn func(ctx AccessContext) RuleResult
}

func NewRuleFunc(id string, fn func(ctx AccessContext) RuleResult) RuleFunc {
	return RuleFunc{id: id, fn: fn}
}

func (f RuleFunc) ID() string {
	return f.id
}

func (f RuleFunc) Evaluate(ctx AccessContext) RuleResult {
	if f.fn == nil {
		return AbstainRuleResult()
	}
	return f.fn(ctx)
}

type AccessContext struct {
	User     User
	Document Document
	Request  AccessRequest
	Region   string
}

type Effect string

const (
	EffectAbstain Effect = "abstain"
	EffectAllow   Effect = "allow"
	EffectDeny    Effect = "deny"
)

type RuleResult struct {
	Effect Effect
	Status Status
	Reason string
}

func evaluateRules(ctx AccessContext, rules []Rule) Decision {
	for _, rule := range rules {
		result := rule.Evaluate(ctx)
		switch result.Effect {
		case EffectAllow:
			return allow(result.Reason)
		case EffectDeny:
			return deny(result.Status, result.Reason)
		}
	}

	return deny(StatusForbidden, "no matching policy")
}

func allow(reason string) Decision {
	return Decision{Allowed: true, Status: StatusAllowed, Reason: reason}
}

func deny(status Status, reason string) Decision {
	return Decision{Allowed: false, Status: status, Reason: reason}
}

func allowRule(reason string) RuleResult {
	return RuleResult{Effect: EffectAllow, Status: StatusAllowed, Reason: reason}
}

func denyRule(status Status, reason string) RuleResult {
	return RuleResult{Effect: EffectDeny, Status: status, Reason: reason}
}

func abstain() RuleResult {
	return RuleResult{Effect: EffectAbstain}
}

func AllowRuleResult(reason string) RuleResult {
	return allowRule(reason)
}

func DenyRuleResult(status Status, reason string) RuleResult {
	return denyRule(status, reason)
}

func AbstainRuleResult() RuleResult {
	return abstain()
}
