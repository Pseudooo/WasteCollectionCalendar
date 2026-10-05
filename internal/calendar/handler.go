package calendar

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Pseudooo/WasteCollectionCalendar/internal/models"
	ics "github.com/arran4/golang-ical"
	"github.com/gin-gonic/gin"
)

type CalendarHandler struct {
	Logger *slog.Logger
}

type CalendarQuery struct {
	Uprn     string `form:"uprn" binding:"required"`
	Postcode string `form:"postcode" binding:"required"`
}

func (h *CalendarHandler) GetCalendar(ctx *gin.Context) {
	logger := h.Logger
	if ctxLogger, exists := ctx.Get("slog_logger"); exists {
		if l, ok := ctxLogger.(*slog.Logger); ok {
			logger = l
		}
	}

	var query CalendarQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var allEvents []models.WasteCollectionEvent
	currentTime := time.Now()
	for i := range 3 {
		evalTime := currentTime.AddDate(0, i, 0)
		events, err := getWasteCollectionEvents(ctx, query.Uprn, query.Postcode, int(evalTime.Month()), evalTime.Year())
		if err != nil {
			logger.Error(
				"Error when calling gov api",
				slog.Any("error", err),
			)
			ctx.AbortWithError(500, err)
			return
		}

		allEvents = append(allEvents, events...)
	}

	calendar := buildCalendarFromWasteCollectionEvents(allEvents)

	ctx.Header("Content-Type", "text/calendar; charset=utf-8")
	ctx.Header("Content-Disposition", `attachment; filename="calendar.ics"`)
	ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")

	ctx.String(http.StatusOK, calendar.Serialize(ics.WithNewLineWindows))
}
