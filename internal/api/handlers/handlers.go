package handlers

import (
	"allaboutapps.dev/aw/go-starter/internal/api"
	"allaboutapps.dev/aw/go-starter/internal/api/handlers/common"
	"allaboutapps.dev/aw/go-starter/internal/api/handlers/projects"
)

func AttachAllRoutes(s *api.Server) {
	// Common routes
	s.Router.Management.GET("/version", common.GetVersion(s))

	// Projects routes
	s.Router.APIV1Projects.GET("", projects.ListProjects(s))
	s.Router.APIV1Projects.GET("/:name/status", projects.GetProjectStatus(s))
	s.Router.APIV1Projects.POST("/:name/deploy", projects.TriggerDeploy(s))
	s.Router.APIV1Projects.GET("/:name/logs", projects.StreamProjectLogs(s))
}
