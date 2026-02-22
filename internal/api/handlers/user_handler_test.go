package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/api/handlers"
	"github.com/misalima/edunex-backend/internal/api/handlers/mocks"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/domain_errors"
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
				err = json.Unmarshal(recorder.Body.Bytes(), &resp)
				Expect(err).To(BeNil())
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
			err = json.Unmarshal(recorder.Body.Bytes(), &resp)
			Expect(err).To(BeNil())
			Expect(resp).To(HaveLen(2))
			Expect(resp[0].(map[string]interface{})["name"]).To(Equal("User 1"))
			Expect(resp[1].(map[string]interface{})["name"]).To(Equal("User 2"))
		})
	})
})
