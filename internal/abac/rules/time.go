package rules

import "github.com/Becks723/go-abac-demo/internal/abac"

type Time struct{}

func (Time) ID() string {
	return "time"
}

func (Time) Evaluate(ctx abac.AccessContext) abac.RuleResult {
	if ctx.Document.StartHour == 0 && ctx.Document.EndHour == 0 {
		return abac.AbstainRuleResult()
	}
	if ctx.Request.Hour >= ctx.Document.StartHour && ctx.Request.Hour < ctx.Document.EndHour {
		return abac.AbstainRuleResult()
	}
	return abac.DenyRuleResult(abac.StatusTimeDenied, "time restricted")
}
