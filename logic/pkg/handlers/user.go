package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"pdm-logic-server/pkg/errors"
	"pdm-logic-server/pkg/models"
	"pdm-logic-server/pkg/services"
	"strconv"
	"strings"
	"time"
)

type UserHandler struct {
	*BaseHandler
}

func NewUserHandler(base *BaseHandler) *UserHandler {
	return &UserHandler{BaseHandler: base}
}

func (h *UserHandler) ValidateVerificationCode(c echo.Context) error {

	var req models.VerificationRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewAppError(http.StatusBadRequest, "Invalid request format", err)
	}

	if err := c.Validate(&req); err != nil {
		return errors.NewAppError(http.StatusBadRequest, "Invalid request data", err)
	}

	if !services.ValidateVerificationCode(h.storage, req.Email, req.VerificationCode) {
		return errors.NewAppError(http.StatusUnauthorized, "Invalid verification code", nil)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Verification successful",
	})
}

func (h *UserHandler) ResendVerificationCode(email, verificationCode string) error {
	from := "hi@demomailtrap.com"
	to := email
	subject := "PDM Notes Registration Code"
	body := "This is the registration code for PDM Notes. "

	if err := services.SendEmail(from, to, subject, body, verificationCode, h.BaseHandler.config.Email.ApiKey); err != nil {
		log.Println("Failed to send email: ", err)
		return errors.NewAppError(http.StatusInternalServerError, "Failed to send email", err)
	}

	return nil
}

// Helper function to get real IP address
func getRealIP(c echo.Context) string {
	// First check CF-Connecting-IP header (most reliable for Cloudflare)
	if ip := c.Request().Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}

	// Then check X-Forwarded-For
	if xff := c.Request().Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			// Get the first IP (client IP) and trim space
			return strings.TrimSpace(ips[0])
		}
	}

	// Then X-Real-IP
	if xrip := c.Request().Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}

	// Finally fall back to RemoteAddr
	remoteAddr := c.Request().RemoteAddr
	if ip, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return ip
	}

	return remoteAddr
}

func (h *UserHandler) verifyTurnstile(token string, clientIP string) (*models.TurnstileResponse, error) {
	formData := url.Values{}
	formData.Set("secret", h.BaseHandler.config.Email.TurnstileSecretKey)
	formData.Set("response", token)
	formData.Set("remoteip", clientIP)

	log.Printf("Verifying turnstile token for IP: %s", clientIP)

	resp, err := http.PostForm(
		"https://challenges.cloudflare.com/turnstile/v0/siteverify",
		formData,
	)
	if err != nil {
		log.Printf("Turnstile verification request failed: %v", err)
		return nil, fmt.Errorf("failed to verify turnstile: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	log.Printf("Turnstile response: %s", string(body))

	var result models.TurnstileResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Failed to decode response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if !result.Success {
		log.Printf("Turnstile verification failed. Errors: %v", result.ErrorCodes)
	}

	return &result, nil
}

func (h *UserHandler) Register(c echo.Context) error {
	ctx := context.Background()

	var req models.SignupRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewAppError(http.StatusBadRequest, "Invalid request format", err)
	}

	if err := c.Validate(&req); err != nil {
		return errors.NewAppError(http.StatusBadRequest, "Invalid request data", err)
	}

	clientIP := getRealIP(c)
	log.Printf("Processing registration request from IP: %s", clientIP)

	result, err := h.verifyTurnstile(req.TurnstileToken, clientIP)
	if err != nil {
		log.Printf("Turnstile verification error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Verification failed",
			"error":   err.Error(),
		})
	}

	if !result.Success {
		errorMsg := "Verification failed"
		if len(result.ErrorCodes) > 0 {
			errorMsg = fmt.Sprintf("Verification failed: %v", result.ErrorCodes)
		}
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": errorMsg,
			"errors":  result.ErrorCodes,
		})
	}

	// Strip whitespace from email
	req.Email = strings.TrimSpace(req.Email)

	// Store the user in the database
	signupInternalRes, err := services.RegisterUser(h.storage, ctx, "", req.Email, req.Password) // Use empty string for name
	if err != nil {
		return c.JSON(http.StatusConflict, map[string]interface{}{
			"message": "Email already exists",
		})
	}

	from := "register@pdm.pw"
	to := req.Email
	subject := "PDM Notes Registration Code"
	body := "This is the registration code for PDM Notes. "

	if err := services.SendEmail(from, to, subject, body, signupInternalRes.VerificationCode, h.BaseHandler.config.Email.ApiKey); err != nil {
		log.Println("Failed to send email: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"userId":  signupInternalRes.UserId,
			"message": "Signup successful, but verification email failed to send, please try again later",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"userId":  signupInternalRes.UserId,
		"message": "Registration successful, check your email for verification code",
	})
}

func (h *UserHandler) Login(c echo.Context) error {
	ctx := context.Background()

	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewAppError(http.StatusBadRequest, "Invalid request format", err)
	}

	if err := c.Validate(&req); err != nil {
		return errors.NewAppError(http.StatusBadRequest, "Invalid request data", err)
	}

	clientIP := getRealIP(c)

	result, err := h.verifyTurnstile(req.TurnstileToken, clientIP)
	if err != nil {
		log.Printf("Signin verification error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Verification failed",
			"error":   err.Error(),
		})
	}

	if !result.Success {
		errorMsg := "Verification failed"
		if len(result.ErrorCodes) > 0 {
			errorMsg = fmt.Sprintf("Signin failed: %v", result.ErrorCodes)
		}
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": errorMsg,
			"errors":  result.ErrorCodes,
		})
	}

	// Strip whitespace from email
	req.Email = strings.TrimSpace(req.Email)

	userId, isValid := services.ValidateUser(h.storage, ctx, req.Email, req.Password)
	if !isValid {
		return errors.NewAppError(http.StatusUnauthorized, "Invalid credentials", nil)
	}

	userinfo, err := services.GetUserInfo(h.storage, ctx, userId)
	if err != nil {
		return errors.NewAppError(http.StatusInternalServerError, "Failed to get user info", err)
	}
	if userinfo.Registered == "0" {
		code, err := services.MakeNewVerificationCode(h.storage, ctx, req.Email)
		if err != nil {
			return errors.NewAppError(http.StatusInternalServerError, "Unverified Email: Failed to make new verification code", err)
		}
		err = h.ResendVerificationCode(req.Email, code)
		if err != nil {
			return errors.NewAppError(http.StatusInternalServerError, "Unverified Email: Failed to resend verification code", err)
		}
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"message": "Unverified Email: new verification email sent, please check",
			"error":   "StatusUnauthorized",
		})
	}

	token, expiration, err := h.authService.GenerateToken(req.Email, userId)
	if err != nil {
		return errors.NewAppError(http.StatusInternalServerError, "Failed to generate token", err)
	}

	h.storage.R.DispatchAddSession(strconv.Itoa(int(userId)), token, expiration)
	if err := h.cacheUserSession(ctx, req.Email, userId, token, time.Unix(expiration, 0)); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"sessionKey": token,
		"expiration": expiration,
		"message":    "Login successful",
	})
}

func (h *UserHandler) cacheUserSession(ctx context.Context, email string, userId uint, token string, expiration time.Time) error {
	// Cache user ID mapping
	if err := h.storage.Ch.HSet(ctx, "userEmail:userId", email, strconv.FormatUint(uint64(userId), 10)); err != nil {
		return errors.NewAppError(http.StatusInternalServerError, "Failed to cache user mapping", err)
	}

	// Cache session token
	key := fmt.Sprintf("user:%d:sessionKey", userId)
	ttl := time.Until(expiration)
	if err := h.storage.Ch.Set(ctx, key, token, ttl); err != nil {
		return errors.NewAppError(http.StatusInternalServerError, "Failed to cache session", err)
	}

	return nil
}

func (h *UserHandler) Logout(c echo.Context) error {
	ctx := context.Background()
	log.Println("Logout request received")

	userEmail := c.Get("email").(string)

	userId, err := h.storage.Ch.HGet(ctx, "userEmail:userId", userEmail)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed, HGet userEmail:userId",
		})
	}
	key := fmt.Sprintf("user:%s:sessionKey", userId)

	sessionKey, err := h.storage.Ch.Get(ctx, key)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed, Get user:%s:sessionKeyy",
		})
	}

	token, err := h.authService.ValidateToken(sessionKey)
	if err != nil {
		log.Println("Token invalid at logout: ", err)
	}

	claims := token.Claims.(jwt.MapClaims)
	parsedExp := time.Unix(int64(claims["exp"].(float64)), 0)
	ttl := time.Until(parsedExp) // Calculate duration until expiration (for Redis TTL)

	// Get blocked list count
	maxCount, err := h.storage.Ch.CountKeys(ctx, fmt.Sprintf("user:%s:blocked:*", userId))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed, CountKeys user:%s:blocked:*",
		})
	}

	// Add session key to blocked list
	blockListKey := fmt.Sprintf("user:%s:blocked:%d", userId, maxCount+1)
	err = h.storage.Ch.Set(ctx, blockListKey, sessionKey, ttl)

	err = h.storage.Ch.Delete(ctx, key)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed, Delete user:%s:sessionKey",
		})
	}

	// Keep the session key in the database
	//h.storage.R.DispatchDeleteSession(userId, sessionKey)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Logout successful",
	})
}

func (h *UserHandler) GetUserInfo(c echo.Context) error {
	ctx := context.Background()
	log.Println("Get user info request received")

	userEmail := c.Get("email").(string)

	userId, err := h.storage.Ch.HGet(ctx, "userEmail:userId", userEmail)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Cache operation failed, HGet userEmail:userId",
		})
	}

	intUserId, err := strconv.Atoi(userId)
	userInfo, err := services.GetUserInfo(h.storage, ctx, uint(intUserId))

	return c.JSON(http.StatusOK, userInfo)
}
