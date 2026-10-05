package calendar

import (
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Pseudooo/WasteCollectionCalendar/internal/models"
	govWasteApi "github.com/Pseudooo/WasteCollectionCalendar/internal/repositories"
	ics "github.com/arran4/golang-ical"
	"github.com/gin-gonic/gin"
)

type CalendarHandler struct {
	Logger     *slog.Logger
	Repository *govWasteApi.GovWasteApiRepository
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

	results := make([][]models.WasteCollectionEvent, 3)
	var wg sync.WaitGroup
	var hasError atomic.Int32

	for i := range 3 {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			if hasError.Load() == 1 {
				return
			}

			evalTime := currentTime.AddDate(0, i, 0)
			events, err := h.Repository.GetWasteCollectionEvents(ctx, query.Uprn, query.Postcode, int(evalTime.Month()), evalTime.Year())
			if err != nil {
				// Mark that an error occurred
				if hasError.CompareAndSwap(0, 1) {
					logger.Error("Error when calling gov api", slog.Any("error", err))
					ctx.AbortWithError(500, err)
				}
				return
			}

			results[index] = events
		}(i)
	}

	wg.Wait()
	if hasError.Load() == 1 {
		return
	}

	for _, events := range results {
		allEvents = append(allEvents, events...)
	}

	calendar := buildCalendarFromWasteCollectionEvents(allEvents)

	ctx.Header("Content-Type", "text/calendar; charset=utf-8")
	ctx.Header("Content-Disposition", `attachment; filename="calendar.ics"`)
	ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")

	ctx.String(http.StatusOK, calendar.Serialize(ics.WithNewLineWindows))
}
