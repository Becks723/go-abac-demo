package abac

import "testing"

func TestCheckAccessExplicitPermission(t *testing.T) {
	svc := newTestService()

	got := svc.CheckAccess(AccessRequest{
		UserID:     "u2",
		DocumentID: "doc1",
		Action:     ActionView,
		Region:     "CN",
		Hour:       10,
	})

	if !got.Allowed {
		t.Fatalf("expected allowed, got %+v", got)
	}
}

func TestCheckAccessSameDepartmentView(t *testing.T) {
	svc := newTestService()

	got := svc.CheckAccess(AccessRequest{
		UserID:     "u1",
		DocumentID: "doc1",
		Action:     ActionView,
		Region:     "CN",
		Hour:       10,
	})

	if !got.Allowed {
		t.Fatalf("expected allowed, got %+v", got)
	}
}

func TestCheckAccessOwnerDraftEdit(t *testing.T) {
	svc := newTestService()

	got := svc.CheckAccess(AccessRequest{
		UserID:     "u1",
		DocumentID: "doc1",
		Action:     ActionEdit,
		Region:     "CN",
		Hour:       10,
	})

	if !got.Allowed {
		t.Fatalf("expected allowed, got %+v", got)
	}
}

func TestCheckAccessLocationDenied(t *testing.T) {
	svc := newTestService()

	got := svc.CheckAccess(AccessRequest{
		UserID:     "u2",
		DocumentID: "doc1",
		Action:     ActionView,
		Region:     "US",
		Hour:       10,
	})

	if got.Allowed || got.Status != StatusLocationDenied {
		t.Fatalf("expected location denial, got %+v", got)
	}
}

func TestCheckAccessTimeDenied(t *testing.T) {
	svc := newTestService()

	got := svc.CheckAccess(AccessRequest{
		UserID:     "u2",
		DocumentID: "doc1",
		Action:     ActionView,
		Region:     "CN",
		Hour:       23,
	})

	if got.Allowed || got.Status != StatusTimeDenied {
		t.Fatalf("expected time denial, got %+v", got)
	}
}

func TestCheckAccessPointsDenied(t *testing.T) {
	svc := newTestService()

	got := svc.CheckAccess(AccessRequest{
		UserID:     "u3",
		DocumentID: "doc1",
		Action:     ActionView,
		Region:     "CN",
		Hour:       10,
	})

	if got.Allowed || got.Status != StatusPaymentDenied {
		t.Fatalf("expected points denial, got %+v", got)
	}
}

func TestHandleEventUpdatesPoints(t *testing.T) {
	store := newTestStore()
	svc := NewServiceWithEnforcer(store, NewEnforcer(defaultTestRules()...))

	published, err := svc.HandleEvent(Event{UserID: "u1", Action: ActionPublish})
	if err != nil {
		t.Fatal(err)
	}
	if published.Delta != -10 || published.Points != 90 {
		t.Fatalf("unexpected publish result: %+v", published)
	}

	invited, err := svc.HandleEvent(Event{UserID: "u1", Action: ActionInvite})
	if err != nil {
		t.Fatal(err)
	}
	if invited.Delta != 5 || invited.Points != 95 {
		t.Fatalf("unexpected invite result: %+v", invited)
	}
}

func TestCheckAccessCanInjectNewRuleWithoutChangingService(t *testing.T) {
	store := newTestStore()
	rules := append([]Rule{
		NewRuleFunc("temporary-business-exception", func(ctx AccessContext) RuleResult {
			if ctx.User.ID == "u3" && ctx.Request.Action == ActionView {
				return allowRule("temporary business exception")
			}
			return abstain()
		}),
	}, defaultTestRules()...)
	svc := NewServiceWithEnforcer(store, NewEnforcer(rules...))

	got := svc.CheckAccess(AccessRequest{
		UserID:     "u3",
		DocumentID: "doc1",
		Action:     ActionView,
		Region:     "CN",
		Hour:       10,
	})

	if !got.Allowed || got.Reason != "temporary business exception" {
		t.Fatalf("expected custom rule to allow request, got %+v", got)
	}
}

func TestEnforcerCanRemoveRuleAtRuntime(t *testing.T) {
	store := newTestStore()
	enforcer := NewEnforcer(defaultTestRules()...)
	svc := NewServiceWithEnforcer(store, enforcer)

	blocked := svc.CheckAccess(AccessRequest{
		UserID:     "u2",
		DocumentID: "doc1",
		Action:     ActionView,
		Region:     "US",
		Hour:       10,
	})
	if blocked.Allowed || blocked.Status != StatusLocationDenied {
		t.Fatalf("expected region rule to block request, got %+v", blocked)
	}

	if !enforcer.RemoveRule("region") {
		t.Fatal("expected region rule to be removed")
	}

	allowed := svc.CheckAccess(AccessRequest{
		UserID:     "u2",
		DocumentID: "doc1",
		Action:     ActionView,
		Region:     "US",
		Hour:       10,
	})
	if !allowed.Allowed {
		t.Fatalf("expected request to pass after removing region rule, got %+v", allowed)
	}
}

func newTestService() *Service {
	return NewServiceWithEnforcer(newTestStore(), NewEnforcer(defaultTestRules()...))
}

func defaultTestRules() []Rule {
	return []Rule{
		RegionRule{},
		TimeRule{},
		PointsRule{},
		ExplicitPermissionRule{},
		SameDepartmentViewRule{},
		OwnerDraftRule{},
	}
}

func newTestStore() *MemoryStore {
	store := NewMemoryStore()
	store.SaveUser(User{ID: "u1", Name: "Alice", Department: "legal", Region: "CN", Points: 100})
	store.SaveUser(User{ID: "u2", Name: "Bob", Department: "legal", Region: "US", Points: 30})
	store.SaveUser(User{ID: "u3", Name: "Carol", Department: "finance", Region: "CN", Points: 5})

	store.SaveDocument(Document{
		ID:         "doc1",
		Title:      "Legal Draft",
		OwnerID:    "u1",
		Department: "legal",
		Status:     StatusDraft,
		AllowedUsers: map[string]Permission{
			"u2": {CanView: true, CanEdit: false},
		},
		AllowedRegions: map[string]bool{"CN": true},
		MinPoints:      10,
		StartHour:      9,
		EndHour:        18,
	})

	store.SaveDocument(Document{
		ID:         "doc2",
		Title:      "Finance Report",
		OwnerID:    "u3",
		Department: "finance",
		Status:     StatusPublished,
		AllowedUsers: map[string]Permission{
			"u1": {CanView: true, CanEdit: true},
		},
		AllowedRegions: map[string]bool{"CN": true, "US": true},
		MinPoints:      0,
		StartHour:      0,
		EndHour:        24,
	})

	return store
}
