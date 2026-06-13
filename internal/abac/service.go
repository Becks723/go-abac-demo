package abac

import "fmt"

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) CheckAccess(req AccessRequest) Decision {
	user, ok := s.store.User(req.UserID)
	if !ok {
		return deny(StatusForbidden, "user not found")
	}

	doc, ok := s.store.Document(req.DocumentID)
	if !ok {
		return deny(StatusForbidden, "document not found")
	}

	region := req.Region
	if region == "" {
		region = user.Region
	}

	if !isRegionAllowed(doc, region) {
		return deny(StatusLocationDenied, "location restricted")
	}

	if !isHourAllowed(doc, req.Hour) {
		return deny(StatusTimeDenied, "time restricted")
	}

	if user.Points < doc.MinPoints {
		return deny(StatusPaymentDenied, "not enough points")
	}

	if isExplicitlyAllowed(doc, user.ID, req.Action) {
		return allow("explicit permission matched")
	}

	if doc.Department == user.Department && req.Action == ActionView {
		return allow("same department can view")
	}

	if doc.OwnerID == user.ID && doc.Status == StatusDraft && canOwnerDraftAction(req.Action) {
		return allow("owner can edit draft and configure permissions")
	}

	return deny(StatusForbidden, "no matching policy")
}

func (s *Service) HandleEvent(event Event) (EventResult, error) {
	if _, ok := s.store.User(event.UserID); !ok {
		return EventResult{}, fmt.Errorf("user not found")
	}

	delta, ok := pointsDelta(event.Action)
	if !ok {
		return EventResult{}, fmt.Errorf("unsupported billable action: %s", event.Action)
	}

	points, ok := s.store.UpdateUserPoints(event.UserID, delta)
	if !ok {
		return EventResult{}, fmt.Errorf("user not found")
	}

	return EventResult{Delta: delta, Points: points}, nil
}

func pointsDelta(action Action) (int, bool) {
	switch action {
	case ActionPublish:
		return -10, true
	case ActionShare, ActionInvite:
		return 5, true
	case ActionClick, ActionBrowse:
		return 0, true
	default:
		return 0, false
	}
}

func isExplicitlyAllowed(doc Document, userID string, action Action) bool {
	permission, ok := doc.AllowedUsers[userID]
	if !ok {
		return false
	}

	switch action {
	case ActionView:
		return permission.CanView || permission.CanEdit
	case ActionEdit:
		return permission.CanEdit
	default:
		return false
	}
}

func isRegionAllowed(doc Document, region string) bool {
	if len(doc.AllowedRegions) == 0 {
		return true
	}
	return doc.AllowedRegions[region]
}

func isHourAllowed(doc Document, hour int) bool {
	if doc.StartHour == 0 && doc.EndHour == 0 {
		return true
	}
	return hour >= doc.StartHour && hour < doc.EndHour
}

func canOwnerDraftAction(action Action) bool {
	return action == ActionView || action == ActionEdit
}

func allow(reason string) Decision {
	return Decision{Allowed: true, Status: StatusAllowed, Reason: reason}
}

func deny(status Status, reason string) Decision {
	return Decision{Allowed: false, Status: status, Reason: reason}
}
