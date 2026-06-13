package rules

import "go-abac-demo/internal/abac"

type OwnerDraft struct{}

func (OwnerDraft) ID() string {
	return "owner-draft"
}

func (OwnerDraft) Evaluate(ctx abac.AccessContext) abac.RuleResult {
	if ctx.Document.OwnerID != ctx.User.ID || ctx.Document.Status != abac.StatusDraft {
		return abac.AbstainRuleResult()
	}
	if ctx.Request.Action == abac.ActionView || ctx.Request.Action == abac.ActionEdit {
		return abac.AllowRuleResult("owner can edit draft and configure permissions")
	}
	return abac.AbstainRuleResult()
}
