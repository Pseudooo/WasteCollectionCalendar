package models

import "time"

type WasteCollectionEvent struct {
	Title string
	Date  time.Time
}
