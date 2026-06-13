package main

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/Becks723/go-abac-demo/internal/abac"
	"github.com/Becks723/go-abac-demo/internal/abac/rules"
)

func main() {
	store := abac.NewMemoryStore()
	initDemoData(store)

	enforcer := abac.NewEnforcer(defaultRules()...)
	svc := abac.NewServiceWithEnforcer(store, enforcer)

	h := server.Default(server.WithHostPorts(":8888"))

	h.POST("/access/check", func(ctx context.Context, c *app.RequestContext) {
		var req accessRequest
		if err := c.BindAndValidate(&req); err != nil {
			c.JSON(consts.StatusBadRequest, response{
				Allowed: false,
				Status:  abac.StatusForbidden,
				Reason:  "invalid request",
			})
			return
		}

		result := svc.CheckAccess(abac.AccessRequest{
			UserID:     req.UserID,
			DocumentID: req.DocumentID,
			Action:     abac.Action(req.Action),
			Region:     req.Region,
			Hour:       req.Hour,
		})
		c.JSON(consts.StatusOK, responseFromDecision(result))
	})

	h.POST("/events", func(ctx context.Context, c *app.RequestContext) {
		var req eventRequest
		if err := c.BindAndValidate(&req); err != nil {
			c.JSON(consts.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}

		result, err := svc.HandleEvent(abac.Event{
			UserID:     req.UserID,
			DocumentID: req.DocumentID,
			Action:     abac.Action(req.Action),
		})
		if err != nil {
			c.JSON(consts.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		c.JSON(consts.StatusOK, map[string]any{
			"user_id": req.UserID,
			"action":  req.Action,
			"delta":   result.Delta,
			"points":  result.Points,
		})
	})

	h.GET("/users/:id/points", func(ctx context.Context, c *app.RequestContext) {
		id := c.Param("id")
		user, ok := store.User(id)
		if !ok {
			c.JSON(consts.StatusNotFound, map[string]string{"error": "user not found"})
			return
		}

		c.JSON(consts.StatusOK, map[string]any{
			"user_id": user.ID,
			"points":  user.Points,
		})
	})

	h.Spin()
}

func defaultRules() []abac.Rule {
	return []abac.Rule{
		rules.Region{},
		rules.Time{},
		rules.Points{},
		rules.ExplicitPermission{},
		rules.SameDepartmentView{},
		rules.OwnerDraft{},
	}
}

func initDemoData(store *abac.MemoryStore) {
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
}

type accessRequest struct {
	UserID     string `json:"user_id" vd:"len($)>0"`
	DocumentID string `json:"document_id" vd:"len($)>0"`
	Action     string `json:"action" vd:"len($)>0"`
	Region     string `json:"region"`
	Hour       int    `json:"hour"`
}

type eventRequest struct {
	UserID     string `json:"user_id" vd:"len($)>0"`
	DocumentID string `json:"document_id"`
	Action     string `json:"action" vd:"len($)>0"`
}

type response struct {
	Allowed bool        `json:"allowed"`
	Status  abac.Status `json:"status"`
	Reason  string      `json:"reason"`
}

func responseFromDecision(d abac.Decision) response {
	return response{
		Allowed: d.Allowed,
		Status:  d.Status,
		Reason:  d.Reason,
	}
}
