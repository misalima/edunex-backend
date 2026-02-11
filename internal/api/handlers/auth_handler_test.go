package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/api/handlers"
	"github.com/misalima/edunex-backend/internal/api/handlers/dto/request"
	"github.com/misalima/edunex-backend/internal/api/handlers/mocks"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/domain_errors"
	"github.com/misalima/edunex-backend/internal/core/util"
)

var _ = Describe("AuthHandler", func() {
	var (
		ctrl         *gomock.Controller
		mockAuthMgr  *mocks.MockAuthManager
		mockUserMgr  *mocks.MockUserManager
		authHandler  *handlers.AuthHandler
		echoInstance *echo.Echo
		recorder     *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockAuthMgr = mocks.NewMockAuthManager(ctrl)
		mockUserMgr = mocks.NewMockUserManager(ctrl)
		authHandler = handlers.NewAuthHandler(mockAuthMgr, mockUserMgr)
		echoInstance = echo.New()
		recorder = httptest.NewRecorder()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Login", func() {
		Context("with valid credentials", func() {
			It("should return success response with tokens", func() {
				// Arrange
				loginReq := request.LoginRequest{
					Email:    "test@example.com",
					Password: "password123",
				}
				reqBody, _ := json.Marshal(loginReq)
				req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
				echoCtx := echoInstance.NewContext(req, recorder)

				expectedUser := &domain.User{
					ID:    uuid.New(),
					Name:  "Test User",
					Email: "test@example.com",
					Role:  "teacher",
				}

				expectedLoginData := &util.LoginResponse{
					AccessToken:  "access-token-123",
					RefreshToken: "refresh-token-456",
					User:         expectedUser,
				}

				mockAuthMgr.EXPECT().
					Login(gomock.Any(), loginReq.Email, loginReq.Password).
					Return(expectedLoginData, nil)

				// Act
				err := authHandler.Login(echoCtx)

				// Assert
				Expect(err).To(BeNil())
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var resp map[string]interface{}
				if err := json.Unmarshal(recorder.Body.Bytes(), &resp)
	if err != nil {
		t.Errorf("Error unmarshalling JSON: %v", err)
}				Expect(resp["token"]).To(Equal("access-token-123"))
				Expect(resp["user"]).ToNot(BeNil())

				// Check cookie
				cookies := recorder.Result().Cookies()
				Expect(cookies).To(HaveLen(1))
				Expect(cookies[0].Name).To(Equal("refresh_token"))
				Expect(cookies[0].Value).To(Equal("refresh-token-456"))
				Expect(cookies[0].HttpOnly).To(BeTrue())
			})
		})

		Context("with invalid request payload", func() {
			It("should return bad request error", func() {
				// Arrange
				req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte("invalid json")))
				req.Header.Set("Content-Type", "application/json")
				echoCtx := echoInstance.NewContext(req, recorder)

				// Act
				err := authHandler.Login(echoCtx)

				// Assert
				Expect(err).To(BeNil())
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))

				var resp map[string]interface{}
				if err := json.Unmarshal(recorder.Body.Bytes(), &resp)
	if err != nil {
		t.Errorf("Error unmarshalling JSON: %v", err)
}				Expect(resp["error"]).To(Equal("invalid request payload"))
			})
		})

		Context("when auth service returns error", func() {
			It("should return unauthorized error", func() {
				// Arrange
				loginReq := request.LoginRequest{
					Email:    "test@example.com",
					Password: "wrongpassword",
				}
				reqBody, _ := json.Marshal(loginReq)
				req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
				echoCtx := echoInstance.NewContext(req, recorder)

				mockAuthMgr.EXPECT().
					Login(gomock.Any(), loginReq.Email, loginReq.Password).
					Return(nil, domain_errors.ErrInvalidCredentials)

				// Act
				err := authHandler.Login(echoCtx)

				// Assert
				Expect(err).To(BeNil())
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("Logout", func() {
		It("should clear refresh token cookie and return success", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
			echoCtx := echoInstance.NewContext(req, recorder)

			// Set a refresh token cookie first
			cookie := &http.Cookie{
				Name:     "refresh_token",
				Value:    "existing-refresh-token",
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
			}
			req.AddCookie(cookie)

			mockAuthMgr.EXPECT().
				Logout(gomock.Any(), "existing-refresh-token").
				Return(nil)

			// Act
			err := authHandler.Logout(echoCtx)

			// Assert
			Expect(err).To(BeNil())
			Expect(recorder.Code).To(Equal(http.StatusNoContent))

			// Check that cookie is cleared
			cookies := recorder.Result().Cookies()
			Expect(cookies).To(HaveLen(1))
			Expect(cookies[0].Name).To(Equal("refresh_token"))
			Expect(cookies[0].Value).To(BeEmpty())
			// MaxAge might be 0 or -1 depending on Echo version
			Expect(cookies[0].MaxAge).To(BeNumerically("<=", 0))
		})
	})

	Describe("Refresh", func() {
		Context("with valid refresh token", func() {
			It("should return new access token", func() {
				// Arrange
				req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
				echoCtx := echoInstance.NewContext(req, recorder)

				// Set refresh token cookie
				cookie := &http.Cookie{
					Name:     "refresh_token",
					Value:    "valid-refresh-token",
					HttpOnly: true,
				}
				req.AddCookie(cookie)

				mockAuthMgr.EXPECT().
					RefreshToken(gomock.Any(), "valid-refresh-token").
					Return("new-access-token-789", nil)

				// Act
				err := authHandler.Refresh(echoCtx)

				// Assert
				Expect(err).To(BeNil())
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var resp map[string]interface{}
				if err := json.Unmarshal(recorder.Body.Bytes(), &resp)
	if err != nil {
		t.Errorf("Error unmarshalling JSON: %v", err)
}				Expect(resp["token"]).To(Equal("new-access-token-789"))
			})
		})

		Context("without refresh token cookie", func() {
			It("should return unauthorized error", func() {
				// Arrange
				req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
				echoCtx := echoInstance.NewContext(req, recorder)

				// Act
				err := authHandler.Refresh(echoCtx)

				// Assert
				Expect(err).To(BeNil())
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})
})
