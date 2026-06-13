package abac

import "fmt"

type Service struct {
	store Store
	rules []Rule
}

func NewService(store Store) *Service {
	return NewServiceWithRules(store, DefaultRules())
}

func NewServiceWithRules(store Store, rules []Rule) *Service {
	return &Service{
		store: store,
		rules: rules,
	}
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

	accessCtx := AccessContext{
		User:     user,
		Document: doc,
		Request:  req,
		Region:   region,
	}
	return EvaluateRules(accessCtx, s.rules)
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
