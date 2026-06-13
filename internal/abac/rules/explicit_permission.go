package rules

import "go-abac-demo/internal/abac"

type ExplicitPermission struct{}

func (ExplicitPermission) ID() string {
	return "explicit-permission"
}

func (ExplicitPermission) Evaluate(ctx abac.AccessContext) abac.RuleResult {
	permission, ok := ctx.Document.AllowedUsers[ctx.User.ID]
	if !ok {
		return abac.AbstainRuleResult()
	}

	switch ctx.Request.Action {
	case abac.ActionView:
		if permission.CanView || permission.CanEdit {
			return abac.AllowRuleResult("explicit permission matched")
		}
	case abac.ActionEdit:
		if permission.CanEdit {
			return abac.AllowRuleResult("explicit permission matched")
		}
	}

	return abac.AbstainRuleResult()
}
