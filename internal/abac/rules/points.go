package rules

import "github.com/Becks723/go-abac-demo/internal/abac"

type Points struct{}

func (Points) ID() string {
	return "points"
}

func (Points) Evaluate(ctx abac.AccessContext) abac.RuleResult {
	if ctx.User.Points >= ctx.Document.MinPoints {
		return abac.AbstainRuleResult()
	}
	return abac.DenyRuleResult(abac.StatusPaymentDenied, "not enough points")
}
