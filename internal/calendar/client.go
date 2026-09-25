package calendar

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Pseudooo/WasteCollectionCalendar/internal/models"
	"github.com/PuerkitoBio/goquery"
)

func getWasteCollectionEvents(uprn string, postcode string, month int, year int) ([]models.WasteCollectionEvent, error) {
	endpoint := "https://ilambassadorformsprod.azurewebsites.net/wastecollectiondays/wastecollectioncalendar"
	data := url.Values{}
	data.Set("Postcode", postcode)
	data.Set("Month", strconv.Itoa(month))
	data.Set("Year", strconv.Itoa(year))
	data.Set("Uprn", uprn)

	client := &http.Client{}
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	events, err := parseWasteCollectionEventsFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func parseWasteCollectionEventsFromReader(reader io.Reader) ([]models.WasteCollectionEvent, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}

	var events []models.WasteCollectionEvent

	doc.Find(".cal-month-box .rc-event-container").Each(func(i int, s *goquery.Selection) {
		eventTitle := strings.TrimSpace(s.Find("span").Text())

		calInnerParent := s.Closest(".cal-inner")

		dateStr, exists := calInnerParent.Find(".day-no").Attr("data-cal-date")
		parsedTime, err := time.Parse("2006-01-02T15:04:05", dateStr)

		if exists && eventTitle != "" && err == nil {
			events = append(events, models.WasteCollectionEvent{
				Title: eventTitle,
				Date:  parsedTime,
			})
		}
	})

	return events, nil
}
