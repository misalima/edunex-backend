package handlers

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/misalima/edunex-backend/internal/api/handlers/dto/request"
	"github.com/misalima/edunex-backend/internal/api/handlers/dto/response"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/interfaces/primary"
	"github.com/misalima/edunex-backend/internal/infra/logger"
	"go.uber.org/zap"
)

type LessonPlanHandler struct {
	lessonPlanService primary.LessonPlanManager
}

func NewLessonPlanHandler(lessonPlanService primary.LessonPlanManager) *LessonPlanHandler {
	return &LessonPlanHandler{lessonPlanService: lessonPlanService}
}

// Create godoc
// @Summary Create lesson plan
// @Description Uploads a lesson plan file and registers its metadata.
// @Tags Lesson Plans
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param title formData string true "Lesson plan title"
// @Param teacher formData string false "Teacher name"
// @Param discipline formData string false "Discipline"
// @Param grade_level formData string false "Grade level"
// @Param file formData file true "Lesson plan file (.pdf or .docx)"
// @Success 201 {object} response.LessonPlanResponse
// @Failure 400 {object} response.ErrorMessageResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /lesson-plans [post]
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
	defer func() {
		if err := file.Close(); err != nil {
			logger.Error("failed to close file", err)
		}
	}()

	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	req := &request.CreateLessonPlanRequest{
		Title:      title,
		Teacher:    c.FormValue("teacher"),
		Discipline: c.FormValue("discipline"),
		GradeLevel: c.FormValue("grade_level"),
	}

	lpDomain := req.ToDomain(userID)

	created, err := h.lessonPlanService.CreateLessonPlan(ctx, lpDomain, file, fileHeader.Filename, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		logger.Log.Error("handler: failed to create lesson plan", zap.Error(err))
		return handleDomainError(c, err)
	}

	return c.JSON(http.StatusCreated, response.FromDomainLessonPlanToResponse(created))
}

// List godoc
// @Summary List lesson plans
// @Description Returns lesson plans for the authenticated user with pagination and signed URLs.
// @Tags Lesson Plans
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Max items per page (1-100, default 20)"
// @Param offset query int false "Number of items to skip (default 0)"
// @Success 200 {object} response.PaginatedLessonPlanResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /lesson-plans [get]
func (h *LessonPlanHandler) List(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	limit := 20
	offset := 0
	if v := c.QueryParam("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 1 && parsed <= 100 {
			limit = parsed
		}
	}
	if v := c.QueryParam("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	params := domain.PaginationParams{Limit: limit, Offset: offset}

	lps, urls, total, err := h.lessonPlanService.ListLessonPlansWithSignedURLs(ctx, userID, params)
	if err != nil {
		return handleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, response.PaginatedLessonPlanResponse{
		Items:  response.FromDomainLessonPlanListWithURLs(lps, urls),
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// GetByID godoc
// @Summary Get lesson plan by ID
// @Description Returns a lesson plan and a signed download URL.
// @Tags Lesson Plans
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lesson plan ID"
// @Success 200 {object} response.LessonPlanResponse
// @Failure 400 {object} response.ErrorMessageResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /lesson-plans/{id} [get]
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

// GetAnalysis godoc
// @Summary Get pedagogical analysis for a lesson plan
// @Description Returns the status and structured pedagogical analysis when ready.
// @Tags Lesson Plans
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lesson plan ID"
// @Success 200 {object} response.LessonPlanAnalysisStatusResponse
// @Failure 400 {object} response.ErrorMessageResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /lesson-plans/{id}/analysis [get]
func (h *LessonPlanHandler) GetAnalysis(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id format"})
	}

	analysisStatus, err := h.lessonPlanService.GetAnalysisStatus(ctx, id)
	if err != nil {
		return handleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, response.FromDomainAnalysisStatus(analysisStatus))
}

// Delete godoc
// @Summary Delete a lesson plan
// @Description Deletes a lesson plan, its file, and associated analyses.
// @Tags Lesson Plans
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lesson plan ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorMessageResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /lesson-plans/{id} [delete]
func (h *LessonPlanHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id format"})
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	if err := h.lessonPlanService.DeleteLessonPlan(ctx, userID, id); err != nil {
		return handleDomainError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
