package main

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"go-abac-demo/internal/abac"
)

func main() {
	store := abac.NewMemoryStore()
	svc := abac.NewService(store)

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
