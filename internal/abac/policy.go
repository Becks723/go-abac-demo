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
		return abstain()
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

type RegionRule struct{}

func (RegionRule) ID() string {
	return "region"
}

func (RegionRule) Evaluate(ctx AccessContext) RuleResult {
	if len(ctx.Document.AllowedRegions) == 0 || ctx.Document.AllowedRegions[ctx.Region] {
		return abstain()
	}
	return denyRule(StatusLocationDenied, "location restricted")
}

type TimeRule struct{}

func (TimeRule) ID() string {
	return "time"
}

func (TimeRule) Evaluate(ctx AccessContext) RuleResult {
	if ctx.Document.StartHour == 0 && ctx.Document.EndHour == 0 {
		return abstain()
	}
	if ctx.Request.Hour >= ctx.Document.StartHour && ctx.Request.Hour < ctx.Document.EndHour {
		return abstain()
	}
	return denyRule(StatusTimeDenied, "time restricted")
}

type PointsRule struct{}

func (PointsRule) ID() string {
	return "points"
}

func (PointsRule) Evaluate(ctx AccessContext) RuleResult {
	if ctx.User.Points >= ctx.Document.MinPoints {
		return abstain()
	}
	return denyRule(StatusPaymentDenied, "not enough points")
}

type ExplicitPermissionRule struct{}

func (ExplicitPermissionRule) ID() string {
	return "explicit-permission"
}

func (ExplicitPermissionRule) Evaluate(ctx AccessContext) RuleResult {
	permission, ok := ctx.Document.AllowedUsers[ctx.User.ID]
	if !ok {
		return abstain()
	}

	switch ctx.Request.Action {
	case ActionView:
		if permission.CanView || permission.CanEdit {
			return allowRule("explicit permission matched")
		}
	case ActionEdit:
		if permission.CanEdit {
			return allowRule("explicit permission matched")
		}
	}

	return abstain()
}

type SameDepartmentViewRule struct{}

func (SameDepartmentViewRule) ID() string {
	return "same-department-view"
}

func (SameDepartmentViewRule) Evaluate(ctx AccessContext) RuleResult {
	if ctx.Document.Department == ctx.User.Department && ctx.Request.Action == ActionView {
		return allowRule("same department can view")
	}
	return abstain()
}

type OwnerDraftRule struct{}

func (OwnerDraftRule) ID() string {
	return "owner-draft"
}

func (OwnerDraftRule) Evaluate(ctx AccessContext) RuleResult {
	if ctx.Document.OwnerID != ctx.User.ID || ctx.Document.Status != StatusDraft {
		return abstain()
	}
	if ctx.Request.Action == ActionView || ctx.Request.Action == ActionEdit {
		return allowRule("owner can edit draft and configure permissions")
	}
	return abstain()
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
