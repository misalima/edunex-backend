package handlers

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/misalima/edunex-backend/internal/api/handlers/dto/response"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/interfaces/iservice"
	"github.com/misalima/edunex-backend/internal/infra/logger"
	"go.uber.org/zap"
)

type LessonPlanHandler struct {
	lessonPlanService iservice.LessonPlanManager
}

func NewLessonPlanHandler(lessonPlanService iservice.LessonPlanManager) *LessonPlanHandler {
	return &LessonPlanHandler{lessonPlanService: lessonPlanService}
}

func (h *LessonPlanHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()

	title := c.FormValue("title")
	if strings.TrimSpace(title) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "title is required"})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file is required"})
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".pdf" && ext != ".docx" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "only .pdf and .docx are allowed"})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to open file"})
	}
	defer file.Close()

	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	lpDomain := &domain.LessonPlan{
		UserID: userID,
		Title:  title,
	}

	created, err := h.lessonPlanService.CreateLessonPlan(ctx, lpDomain, file, fileHeader.Filename, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		logger.Log.Error("handler: failed to create lesson plan", zap.Error(err))
		return handleDomainError(c, err)
	}

	return c.JSON(http.StatusCreated, response.FromDomainLessonPlanToResponse(created))
}

func (h *LessonPlanHandler) List(c echo.Context) error {
	ctx := c.Request().Context()

	lps, urls, err := h.lessonPlanService.ListLessonPlansWithSignedURLs(ctx)
	if err != nil {
		return handleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, response.FromDomainLessonPlanListWithURLs(lps, urls))
}

func (h *LessonPlanHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id format"})
	}

	lp, signedURL, err := h.lessonPlanService.GetLessonPlanWithSignedURL(ctx, id)
	if err != nil {
		return handleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, response.FromDomainLessonPlanWithURL(lp, signedURL))
}
