package abac_test

import (
	"testing"

	"github.com/Becks723/go-abac-demo/internal/abac"
	"github.com/Becks723/go-abac-demo/internal/abac/rules"
)

func TestCheckAccessExplicitPermission(t *testing.T) {
	svc := newTestService()

	got := svc.CheckAccess(abac.AccessRequest{
		UserID:     "u2",
		DocumentID: "doc1",
		Action:     abac.ActionView,
		Region:     "CN",
		Hour:       10,
	})

	if !got.Allowed {
		t.Fatalf("expected allowed, got %+v", got)
	}
}

func TestCheckAccessSameDepartmentView(t *testing.T) {
	svc := newTestService()

	got := svc.CheckAccess(abac.AccessRequest{
		UserID:     "u1",
		DocumentID: "doc1",
		Action:     abac.ActionView,
		Region:     "CN",
		Hour:       10,
	})

	if !got.Allowed {
		t.Fatalf("expected allowed, got %+v", got)
	}
}

func TestCheckAccessOwnerDraftEdit(t *testing.T) {
	svc := newTestService()

	got := svc.CheckAccess(abac.AccessRequest{
		UserID:     "u1",
		DocumentID: "doc1",
		Action:     abac.ActionEdit,
		Region:     "CN",
		Hour:       10,
	})

	if !got.Allowed {
		t.Fatalf("expected allowed, got %+v", got)
	}
}

func TestCheckAccessLocationDenied(t *testing.T) {
	svc := newTestService()

	got := svc.CheckAccess(abac.AccessRequest{
		UserID:     "u2",
		DocumentID: "doc1",
		Action:     abac.ActionView,
		Region:     "US",
		Hour:       10,
	})

	if got.Allowed || got.Status != abac.StatusLocationDenied {
		t.Fatalf("expected location denial, got %+v", got)
	}
}

func TestCheckAccessTimeDenied(t *testing.T) {
	svc := newTestService()

	got := svc.CheckAccess(abac.AccessRequest{
		UserID:     "u2",
		DocumentID: "doc1",
		Action:     abac.ActionView,
		Region:     "CN",
		Hour:       23,
	})

	if got.Allowed || got.Status != abac.StatusTimeDenied {
		t.Fatalf("expected time denial, got %+v", got)
	}
}

func TestCheckAccessPointsDenied(t *testing.T) {
	svc := newTestService()

	got := svc.CheckAccess(abac.AccessRequest{
		UserID:     "u3",
		DocumentID: "doc1",
		Action:     abac.ActionView,
		Region:     "CN",
		Hour:       10,
	})

	if got.Allowed || got.Status != abac.StatusPaymentDenied {
		t.Fatalf("expected points denial, got %+v", got)
	}
}

func TestHandleEventUpdatesPoints(t *testing.T) {
	store := newTestStore()
	svc := abac.NewServiceWithEnforcer(store, abac.NewEnforcer(defaultTestRules()...))

	published, err := svc.HandleEvent(abac.Event{UserID: "u1", Action: abac.ActionPublish})
	if err != nil {
		t.Fatal(err)
	}
	if published.Delta != -10 || published.Points != 90 {
		t.Fatalf("unexpected publish result: %+v", published)
	}

	invited, err := svc.HandleEvent(abac.Event{UserID: "u1", Action: abac.ActionInvite})
	if err != nil {
		t.Fatal(err)
	}
	if invited.Delta != 5 || invited.Points != 95 {
		t.Fatalf("unexpected invite result: %+v", invited)
	}
}

func TestCheckAccessCanInjectNewRuleWithoutChangingService(t *testing.T) {
	store := newTestStore()
	ruleSet := append([]abac.Rule{
		abac.NewRuleFunc("temporary-business-exception", func(ctx abac.AccessContext) abac.RuleResult {
			if ctx.User.ID == "u3" && ctx.Request.Action == abac.ActionView {
				return abac.AllowRuleResult("temporary business exception")
			}
			return abac.AbstainRuleResult()
		}),
	}, defaultTestRules()...)
	svc := abac.NewServiceWithEnforcer(store, abac.NewEnforcer(ruleSet...))

	got := svc.CheckAccess(abac.AccessRequest{
		UserID:     "u3",
		DocumentID: "doc1",
		Action:     abac.ActionView,
		Region:     "CN",
		Hour:       10,
	})

	if !got.Allowed || got.Reason != "temporary business exception" {
		t.Fatalf("expected custom rule to allow request, got %+v", got)
	}
}

func TestEnforcerCanRemoveRuleAtRuntime(t *testing.T) {
	store := newTestStore()
	enforcer := abac.NewEnforcer(defaultTestRules()...)
	svc := abac.NewServiceWithEnforcer(store, enforcer)

	blocked := svc.CheckAccess(abac.AccessRequest{
		UserID:     "u2",
		DocumentID: "doc1",
		Action:     abac.ActionView,
		Region:     "US",
		Hour:       10,
	})
	if blocked.Allowed || blocked.Status != abac.StatusLocationDenied {
		t.Fatalf("expected region rule to block request, got %+v", blocked)
	}

	if !enforcer.RemoveRule("region") {
		t.Fatal("expected region rule to be removed")
	}

	allowed := svc.CheckAccess(abac.AccessRequest{
		UserID:     "u2",
		DocumentID: "doc1",
		Action:     abac.ActionView,
		Region:     "US",
		Hour:       10,
	})
	if !allowed.Allowed {
		t.Fatalf("expected request to pass after removing region rule, got %+v", allowed)
	}
}

func newTestService() *abac.Service {
	return abac.NewServiceWithEnforcer(newTestStore(), abac.NewEnforcer(defaultTestRules()...))
}

func defaultTestRules() []abac.Rule {
	return []abac.Rule{
		rules.Region{},
		rules.Time{},
		rules.Points{},
		rules.ExplicitPermission{},
		rules.SameDepartmentView{},
		rules.OwnerDraft{},
	}
}

func newTestStore() *abac.MemoryStore {
	store := abac.NewMemoryStore()
	store.SaveUser(abac.User{ID: "u1", Name: "Alice", Department: "legal", Region: "CN", Points: 100})
	store.SaveUser(abac.User{ID: "u2", Name: "Bob", Department: "legal", Region: "US", Points: 30})
	store.SaveUser(abac.User{ID: "u3", Name: "Carol", Department: "finance", Region: "CN", Points: 5})

	store.SaveDocument(abac.Document{
		ID:         "doc1",
		Title:      "Legal Draft",
		OwnerID:    "u1",
		Department: "legal",
		Status:     abac.StatusDraft,
		AllowedUsers: map[string]abac.Permission{
			"u2": {CanView: true, CanEdit: false},
		},
		AllowedRegions: map[string]bool{"CN": true},
		MinPoints:      10,
		StartHour:      9,
		EndHour:        18,
	})

	store.SaveDocument(abac.Document{
		ID:         "doc2",
		Title:      "Finance Report",
		OwnerID:    "u3",
		Department: "finance",
		Status:     abac.StatusPublished,
		AllowedUsers: map[string]abac.Permission{
			"u1": {CanView: true, CanEdit: true},
		},
		AllowedRegions: map[string]bool{"CN": true, "US": true},
		MinPoints:      0,
		StartHour:      0,
		EndHour:        24,
	})

	return store
}
