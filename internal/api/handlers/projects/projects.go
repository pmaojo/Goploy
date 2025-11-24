package projects

import (
	"context"
	"fmt"
	"net/http"

	"allaboutapps.dev/aw/go-starter/internal/api"
	"allaboutapps.dev/aw/go-starter/internal/config"
	"github.com/labstack/echo/v4"
)

func ListProjects(s *api.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		projects := make([]string, len(s.GoployConfig.Projects))
		for i, p := range s.GoployConfig.Projects {
			projects[i] = p.Name
		}
		return c.JSON(http.StatusOK, echo.Map{"projects": projects})
	}
}

func GetProjectStatus(s *api.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		projectName := c.Param("name")
		var project *config.Project
		for _, p := range s.GoployConfig.Projects {
			if p.Name == projectName {
				project = &p
				break
			}
		}

		if project == nil {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Project not found"})
		}

		status, err := s.Deployment.GetStatus(c.Request().Context(), *project)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, status)
	}
}

type TriggerDeployRequest struct {
	Ref string `json:"ref"`
}

func TriggerDeploy(s *api.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		projectName := c.Param("name")
		var project *config.Project
		for _, p := range s.GoployConfig.Projects {
			if p.Name == projectName {
				project = &p
				break
			}
		}

		if project == nil {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Project not found"})
		}

		var req TriggerDeployRequest
		if err := c.Bind(&req); err != nil {
			// Optional body, ignore error if empty but check if malformed
		}

		c.Response().Header().Set(echo.HeaderContentType, "text/plain")
		c.Response().WriteHeader(http.StatusOK)

		writer := c.Response()

		// Helper to flush if possible
		flush := func() {
			c.Response().Flush()
		}

		fmt.Fprintf(writer, "Starting deployment for %s (ref: %s)...\n", project.Name, req.Ref)
		flush()

		err := s.Deployment.Deploy(*project, writer, req.Ref)
		if err != nil {
			fmt.Fprintf(writer, "\nDeployment failed: %v\n", err)
			return nil // We already sent 200 OK and started streaming
		}

		fmt.Fprintf(writer, "\nDeployment finished successfully.\n")
		flush()

		return nil
	}
}

func StreamProjectLogs(s *api.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		projectName := c.Param("name")
		var project *config.Project
		for _, p := range s.GoployConfig.Projects {
			if p.Name == projectName {
				project = &p
				break
			}
		}

		if project == nil {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "Project not found"})
		}

		c.Response().Header().Set(echo.HeaderContentType, "text/plain")
		c.Response().WriteHeader(http.StatusOK)

		writer := c.Response()

		// Set a context timeout for the stream if needed, or rely on client disconnect
		// We use request context which cancels on disconnect
		ctx := c.Request().Context()

		err := s.Deployment.StreamLogs(ctx, *project, writer)
		if err != nil {
			// If context is canceled, it's normal disconnect
			if ctx.Err() == context.Canceled {
				return nil
			}
			fmt.Fprintf(writer, "\nLog streaming error: %v\n", err)
		}

		return nil
	}
}
