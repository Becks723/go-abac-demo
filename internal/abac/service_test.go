package abac

import "testing"

func TestCheckAccessExplicitPermission(t *testing.T) {
	svc := NewService(NewMemoryStore())

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
	svc := NewService(NewMemoryStore())

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
	svc := NewService(NewMemoryStore())

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
	svc := NewService(NewMemoryStore())

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
	svc := NewService(NewMemoryStore())

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
	svc := NewService(NewMemoryStore())

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
	store := NewMemoryStore()
	svc := NewService(store)

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
