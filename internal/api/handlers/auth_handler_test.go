package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"fmt"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/api/handlers"
	"github.com/misalima/edunex-backend/internal/api/handlers/dto/request"
	"github.com/misalima/edunex-backend/internal/api/handlers/mocks"
	"github.com/misalima/edunex-backend/internal/core/domain"
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
				if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
					Fail(fmt.Sprintf("Error unmarshalling JSON: %v", err))
				}
				Expect(resp["token"].(string)).To(Equal("access-token-123"))
			})
		})
	})
})