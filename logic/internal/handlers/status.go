// handlers/status.go
package handlers

import (
	"bytes"
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"html/template"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type CustomRenderer struct {
	templates *template.Template
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

func getSystemMetrics() (string, error) {
	cmd := exec.Command("free", "-h")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func getDockerStatus() (string, error) {
	cmd := exec.Command("docker", "ps")
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

	// Create combined template data
	templateData := TemplateData{
		Nonce:  nonce,
		Status: data.(StatusData),
	}

	return r.templates.ExecuteTemplate(w, name, templateData)
}

func SetupRenderer(e *echo.Echo) {
	cwd, err := os.Getwd()
	if err != nil {
		e.Logger.Fatal("Unable to determine the current working directory")
	}

	templatePath := filepath.Join(cwd, "templates", "status.html")
	templates := template.Must(template.ParseFiles(templatePath))

	renderer := &CustomRenderer{
		templates: templates,
	}
	e.Renderer = renderer
}

func StatusHandler(c echo.Context) error {
	// Get password from .env
	expectedPassword := os.Getenv("STATUS_PASSWORD")
	if expectedPassword == "" {
		return echo.NewHTTPError(500, "Status password not configured")
	}

	// Get password from URL
	urlPath := c.Request().URL.Path
	parts := strings.Split(urlPath, "/")
	if len(parts) != 3 || parts[2] != expectedPassword {
		return echo.NewHTTPError(403, "Unauthorized")
	}

	metrics, err := getSystemMetrics()
	if err != nil {
		metrics = "Status 500: Error getting system metrics"
		//return echo.NewHTTPError(500, "Error getting system metrics")
	}

	dockerStatus, err := getDockerStatus()
	if err != nil {
		dockerStatus = "Status 500: Error getting Docker status"
		//return echo.NewHTTPError(500, "Error getting Docker status")
	}

	pm2Status, err := getPM2Status()
	if err != nil {
		pm2Status = "Status 500: Error getting PM2 status"
		//return echo.NewHTTPError(500, "Error getting PM2 status")
	}
	pm2Logs, err := getPM2Logs()
	if err != nil {
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
