package rules

import "go-abac-demo/internal/abac"

type Region struct{}

func (Region) ID() string {
	return "region"
}

func (Region) Evaluate(ctx abac.AccessContext) abac.RuleResult {
	if len(ctx.Document.AllowedRegions) == 0 || ctx.Document.AllowedRegions[ctx.Region] {
		return abac.AbstainRuleResult()
	}
	return abac.DenyRuleResult(abac.StatusLocationDenied, "location restricted")
}
