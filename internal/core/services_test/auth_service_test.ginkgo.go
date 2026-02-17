package services_test

import (
	"errors"
	"context"
	"github.com/misalima/edunex-backend/internal/core/services"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AuthService", func() {
	var (
		authService *services.AuthService
		mockAuthRepo *MockAuthRepo
		mockUserRepo *MockUserRepo
		mockJWTManager *MockJWTManager
		ctx context.Context
	)

	BeforeEach(func() {
		mockAuthRepo = &MockAuthRepo{}
		mockUserRepo = &MockUserRepo{}
		mockJWTManager = &MockJWTManager{}
		authService = services.NewAuthService(mockAuthRepo, mockUserRepo, mockJWTManager)
		ctx = context.TODO()
	})

	Context("Login", func() {
		It("should succeed with valid credentials", func() {
			// Arrange
			mockAuthRepo.On("Login", ctx, "validUser", "validPass").Return(true, nil)

			// Act
			success, err := authService.Login(ctx, "validUser", "validPass")

			// Assert
			Expect(err).To(BeNil())
			Expect(success).To(BeTrue())
			mockAuthRepo.AssertCalled(GinkgoT(), "Login", ctx, "validUser", "validPass")
		})

		It("should fail with invalid credentials", func() {
			// Arrange
			mockAuthRepo.On("Login", ctx, "invalidUser", "invalidPass").Return(false, errors.New("invalid credentials"))

			// Act
			success, err := authService.Login(ctx, "invalidUser", "invalidPass")

			// Assert
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("invalid credentials"))
			Expect(success).To(BeFalse())
			mockAuthRepo.AssertCalled(GinkgoT(), "Login", ctx, "invalidUser", "invalidPass")
		})
	})
})