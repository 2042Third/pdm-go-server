package handlers

import (
	"bytes"
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"html/template"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type CustomRenderer struct {
	template *template.Template
}

type TemplateData struct {
	Nonce  string
	Status StatusData
}

type StatusData struct {
	SystemMetrics string
	DockerStatus  string
	PM2Status     string
	PM2Logs       string
}

type StatusHandler struct {
	*BaseHandler
	statusPassword string
}

func NewStatusHandler(base *BaseHandler, password string) *StatusHandler {
	return &StatusHandler{
		BaseHandler:    base,
		statusPassword: password,
	}
}

func getSystemMetrics() (string, error) {
	cmd := exec.Command("free", "-h")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func getDockerStatus() (string, error) {
	cmd := exec.Command("podman", "ps")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func getPM2Status() (string, error) {
	cmd := exec.Command("pm2", "list")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
func getPM2Logs() (string, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use pm2 logs --lines 100 --nostream to get recent logs
	cmd := exec.CommandContext(ctx, "pm2", "logs", "--lines", "100", "--nostream")

	// Create a buffer to capture output
	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer

	// Run the command
	err := cmd.Run()
	if err != nil {
		// If it's a timeout error, return what we got so far
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return outBuffer.String(), nil
		}
		return "", err
	}

	return outBuffer.String(), nil
}

func (r *CustomRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Get nonce from context
	nonce := c.Get("nonce").(string)
	if nonce == "" {
		return errors.New("nonce not found in context")
	}

	// Create combined template data
	templateData := TemplateData{
		Nonce:  nonce,
		Status: data.(StatusData),
	}

	return r.template.ExecuteTemplate(w, name, templateData)
}

func (s *StatusHandler) SetupRenderer(e *echo.Echo, wd string) {
	templatePath := filepath.Join(wd, "/templates", "status.html")
	templates := template.Must(template.ParseFiles(templatePath))

	renderer := &CustomRenderer{
		template: templates,
	}
	e.Renderer = renderer
}

func (s *StatusHandler) StatusHandlerFunc(c echo.Context) error {

	// Get password from URL
	urlPath := c.Request().URL.Path
	parts := strings.Split(urlPath, "/")
	if len(parts) != 3 || parts[2] != s.statusPassword {
		return echo.NewHTTPError(403, "Unauthorized")
	}

	metrics, err := getSystemMetrics()
	if err != nil {
		s.log.WithError(err).Errorf("Status 500, Error getting system metrics: %s", err.Error())

		metrics = "Status 500, Error getting system metrics"
		//return echo.NewHTTPError(500, "Error getting system metrics")
	}

	dockerStatus, err := getDockerStatus()
	if err != nil {
		s.log.WithError(err).Errorf("Status 500, Error getting Docker status: %s", err.Error())
		dockerStatus = "Status 500: Error getting Docker status"
		//return echo.NewHTTPError(500, "Error getting Docker status")
	}

	pm2Status, err := getPM2Status()
	if err != nil {
		pm2Status = "Status 500: Error getting PM2 status"
		s.log.WithError(err).Errorf("Status 500, Error getting PM2 status: %s", err.Error())

		//return echo.NewHTTPError(500, "Error getting PM2 status")
	}
	pm2Logs, err := getPM2Logs()
	if err != nil {
		s.log.WithError(err).Errorf("Status 500, EError getting PM2 status: %s", err.Error())
		pm2Status = "Status 500: Error getting PM2 status"
		//return echo.NewHTTPError(500, "Error getting PM2 status")
	}

	data := StatusData{
		SystemMetrics: metrics,
		DockerStatus:  dockerStatus,
		PM2Status:     pm2Status,
		PM2Logs:       pm2Logs,
	}

	return c.Render(200, "status.html", data)
}
