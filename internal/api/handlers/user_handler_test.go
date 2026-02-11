package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/labstack/echo/v4"

	"github.com/misalima/edunex-backend/internal/api/handlers"
	"github.com/misalima/edunex-backend/internal/api/handlers/dto/request"
	"github.com/misalima/edunex-backend/internal/api/handlers/mocks"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/domain_errors"
	"github.com/google/uuid"
)

var _ = Describe("UserHandler", func() {
	var (
		ctrl         *gomock.Controller
		mockUserMgr  *mocks.MockUserManager
		userHandler  *handlers.UserHandler
		echoInstance *echo.Echo
		recorder     *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockUserMgr = mocks.NewMockUserManager(ctrl)
		userHandler = handlers.NewUserHandler(mockUserMgr)
		echoInstance = echo.New()
		recorder = httptest.NewRecorder()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("CreateUser", func() {
		Context("with valid user data", func() {
			It("should create user and return created response", func() {
				// Arrange
				createReq := request.CreateUserRequest{
					Name:     "Test User",
					Email:    "test@example.com",
					Password: "password123",
				}
				reqBody, _ := json.Marshal(createReq)
				req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
				echoCtx := echoInstance.NewContext(req, recorder)

				expectedUser := &domain.User{
					ID:       uuid.New(),
					Name:     "Test User",
					Email:    "test@example.com",
					Role:     "teacher",
					Created:  time.Now(),
					Updated:  time.Now(),
				}

				mockUserMgr.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx interface{}, user *domain.User) (*domain.User, error) {
						user.ID = expectedUser.ID
						user.Role = "teacher"
						user.Created = expectedUser.Created
						user.Updated = expectedUser.Updated
						return user, nil
					})

				// Act
				err := userHandler.CreateUser(echoCtx)

				// Assert
				Expect(err).To(BeNil())
				Expect(recorder.Code).To(Equal(http.StatusCreated))

				var resp map[string]interface{}
				json.Unmarshal(recorder.Body.Bytes(), &resp)
				Expect(resp["name"]).To(Equal("Test User"))
				Expect(resp["email"]).To(Equal("test@example.com"))
				Expect(resp["role"]).To(Equal("teacher"))
			})
		})

		Context("with invalid request payload", func() {
			It("should return bad request error", func() {
				// Arrange
				req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader([]byte("invalid json")))
				req.Header.Set("Content-Type", "application/json")
				echoCtx := echoInstance.NewContext(req, recorder)

				// Act
				err := userHandler.CreateUser(echoCtx)

				// Assert
				Expect(err).To(BeNil())
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("with validation errors", func() {
			It("should return bad request error for missing name", func() {
				// Arrange
				createReq := request.CreateUserRequest{
					Email:    "test@example.com",
					Password: "password123",
					// Name is missing
				}
				reqBody, _ := json.Marshal(createReq)
				req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
				echoCtx := echoInstance.NewContext(req, recorder)

				// Act
				err := userHandler.CreateUser(echoCtx)

				// Assert
				Expect(err).To(BeNil())
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("GetUserByID", func() {
		Context("with valid user ID", func() {
			It("should return user data", func() {
				// Arrange
				userID := uuid.New()
				req := httptest.NewRequest(http.MethodGet, "/api/users/"+userID.String(), nil)
				echoCtx := echoInstance.NewContext(req, recorder)
				echoCtx.SetParamNames("id")
				echoCtx.SetParamValues(userID.String())

				expectedUser := &domain.User{
					ID:    userID,
					Name:  "Test User",
					Email: "test@example.com",
					Role:  "teacher",
				}

				mockUserMgr.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(expectedUser, nil)

				// Act
				err := userHandler.GetUserByID(echoCtx)

				// Assert
				Expect(err).To(BeNil())
				Expect(recorder.Code).To(Equal(http.StatusOK))

				var resp map[string]interface{}
				json.Unmarshal(recorder.Body.Bytes(), &resp)
				Expect(resp["id"]).To(Equal(userID.String()))
				Expect(resp["name"]).To(Equal("Test User"))
				Expect(resp["email"]).To(Equal("test@example.com"))
			})
		})

		Context("with invalid user ID format", func() {
			It("should return bad request error", func() {
				// Arrange
				req := httptest.NewRequest(http.MethodGet, "/api/users/invalid-uuid", nil)
				echoCtx := echoInstance.NewContext(req, recorder)
				echoCtx.SetParamNames("id")
				echoCtx.SetParamValues("invalid-uuid")

				// Act
				err := userHandler.GetUserByID(echoCtx)

				// Assert
				Expect(err).To(BeNil())
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when user is not found", func() {
			It("should return not found error", func() {
				// Arrange
				userID := uuid.New()
				req := httptest.NewRequest(http.MethodGet, "/api/users/"+userID.String(), nil)
				echoCtx := echoInstance.NewContext(req, recorder)
				echoCtx.SetParamNames("id")
				echoCtx.SetParamValues(userID.String())

				mockUserMgr.EXPECT().
					GetUserByID(gomock.Any(), userID).
					Return(nil, domain_errors.ErrNotFound)

				// Act
				err := userHandler.GetUserByID(echoCtx)

				// Assert
				Expect(err).To(BeNil())
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
			})
		})
	})

	Describe("ListUsers", func() {
		It("should return list of users", func() {
			// Arrange
			req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
			echoCtx := echoInstance.NewContext(req, recorder)

			expectedUsers := []*domain.User{
				{
					ID:    uuid.New(),
					Name:  "User 1",
					Email: "user1@example.com",
					Role:  "teacher",
				},
				{
					ID:    uuid.New(),
					Name:  "User 2",
					Email: "user2@example.com",
					Role:  "student",
				},
			}

			mockUserMgr.EXPECT().
				ListUsers(gomock.Any()).
				Return(expectedUsers, nil)

			// Act
			err := userHandler.ListUsers(echoCtx)

			// Assert
			Expect(err).To(BeNil())
			Expect(recorder.Code).To(Equal(http.StatusOK))

			var resp []interface{}
			json.Unmarshal(recorder.Body.Bytes(), &resp)
			Expect(resp).To(HaveLen(2))
			Expect(resp[0].(map[string]interface{})["name"]).To(Equal("User 1"))
			Expect(resp[1].(map[string]interface{})["name"]).To(Equal("User 2"))
		})
	})
})