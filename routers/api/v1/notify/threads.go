// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package notify

import (
	"fmt"
	"net/http"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/context"
)

// GetThread get notification by ID
func GetThread(ctx *context.APIContext) {
	// swagger:operation GET /notifications/threads/{id} notification notifyGetThread
	// ---
	// summary: Get notification thread by ID
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: id
	//   in: path
	//   description: id of notification thread
	//   type: string
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/NotificationThread"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"

	n := getThread(ctx)
	if n == nil {
		return
	}
	if err := n.LoadAttributes(); err != nil {
		ctx.InternalServerError(err)
		return
	}

	ctx.JSON(http.StatusOK, n.APIFormat())
}

// ReadThread mark notification as read by ID
func ReadThread(ctx *context.APIContext) {
	// swagger:operation PATCH /notifications/threads/{id} notification notifyReadThread
	// ---
	// summary: Mark notification thread as read by ID
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: id
	//   in: path
	//   description: id of notification thread
	//   type: string
	//   required: true
	// responses:
	//   "205":
	//     "$ref": "#/responses/empty"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"

	n := getThread(ctx)
	if n == nil {
		return
	}

	err := models.SetNotificationStatus(n.ID, ctx.User, models.NotificationStatusRead)
	if err != nil {
		ctx.InternalServerError(err)
		return
	}
	ctx.Status(http.StatusResetContent)
}

func getThread(ctx *context.APIContext) *models.Notification {
	n, err := models.GetNotificationByID(ctx.ParamsInt64(":id"))
	if err != nil {
		if models.IsErrNotExist(err) {
			ctx.Error(http.StatusNotFound, "GetNotificationByID", err)
		} else {
			ctx.InternalServerError(err)
		}
		return nil
	}
	if n.UserID != ctx.User.ID && !ctx.User.IsAdmin {
		ctx.Error(http.StatusForbidden, "GetNotificationByID", fmt.Errorf("only user itself and admin are allowed to read/change this thread %d", n.ID))
		return nil
	}
	return n
}
