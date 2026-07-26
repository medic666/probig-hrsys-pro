package leave_account

import (
	"github.com/gin-gonic/gin"
	"probig/internal/pkg/middleware"
)

func RegisterRoutes(r *gin.RouterGroup) {
	InitDB()

	events := r.Group("/leave-account-events")
	{
		events.GET("", middleware.PermissionMiddleware("leave_account:read"), ListEventsHandler)
		events.GET("/:id", middleware.PermissionMiddleware("leave_account:read"), GetEventHandler)
		events.POST("", middleware.PermissionMiddleware("leave_account:write"), CreateEventHandler)
		events.PUT("/:id", middleware.PermissionMiddleware("leave_account:write"), UpdateEventHandler)
		events.DELETE("/:id", middleware.PermissionMiddleware("leave_account:delete"), DeleteEventHandler)
		events.GET("/trash", middleware.PermissionMiddleware("leave_account:read"), ListTrashHandler)
		events.PUT("/:id/restore", middleware.PermissionMiddleware("leave_account:write"), RestoreEventHandler)
	}

	balances := r.Group("/leave-account-balances")
	{
		balances.GET("", middleware.PermissionMiddleware("leave_account:read"), ListBalancesHandler)
		balances.GET("/:personId/:leaveType/detail", middleware.PermissionMiddleware("leave_account:read"), GetBalanceDetailHandler)
	}

	carryover := r.Group("/leave-account-carryover")
	{
		carryover.POST("/execute", middleware.PermissionMiddleware("leave_account:carryover"), ExecuteCarryoverHandler)
		carryover.POST("/cancel/:batchId", middleware.PermissionMiddleware("leave_account:carryover"), CancelBatchHandler)
	}

	batches := r.Group("/leave-account-batches")
	{
		batches.GET("", middleware.PermissionMiddleware("leave_account:read"), ListBatchesHandler)
	}
}
