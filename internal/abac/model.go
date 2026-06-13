package abac

type Action string

const (
	ActionView    Action = "view"
	ActionEdit    Action = "edit"
	ActionPublish Action = "publish"
	ActionShare   Action = "share"
	ActionInvite  Action = "invite"
	ActionClick   Action = "click"
	ActionBrowse  Action = "browse"
)

type DocumentStatus string

const (
	StatusDraft     DocumentStatus = "draft"
	StatusPublished DocumentStatus = "published"
)

type User struct {
	ID         string
	Name       string
	Department string
	Region     string
	Points     int
}

type Document struct {
	ID             string
	Title          string
	OwnerID        string
	Department     string
	Status         DocumentStatus
	AllowedUsers   map[string]Permission
	AllowedRegions map[string]bool
	MinPoints      int
	StartHour      int
	EndHour        int
}

type Permission struct {
	CanView bool
	CanEdit bool
}

type AccessRequest struct {
	UserID     string
	DocumentID string
	Action     Action
	Region     string
	Hour       int
}

type Event struct {
	UserID     string
	DocumentID string
	Action     Action
}

type Status string

const (
	StatusAllowed        Status = "allowed"
	StatusForbidden      Status = "forbidden"
	StatusLocationDenied Status = "location_restricted"
	StatusTimeDenied     Status = "time_restricted"
	StatusPaymentDenied  Status = "points_required"
)

type Decision struct {
	Allowed bool
	Status  Status
	Reason  string
}

type EventResult struct {
	Delta  int
	Points int
}
