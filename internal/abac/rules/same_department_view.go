package rules

import "github.com/Becks723/go-abac-demo/internal/abac"

type SameDepartmentView struct{}

func (SameDepartmentView) ID() string {
	return "same-department-view"
}

func (SameDepartmentView) Evaluate(ctx abac.AccessContext) abac.RuleResult {
	if ctx.Document.Department == ctx.User.Department && ctx.Request.Action == abac.ActionView {
		return abac.AllowRuleResult("same department can view")
	}
	return abac.AbstainRuleResult()
}
