package data

import "time"

// if we does not include the json annotation for the struct, the default struct value will
// the json keys. like "ID", "CreatedAt" etc.
// but if we include the struct annotation, then that will be the json keys.
// like ID int64 `json:"id"`
// Also note that the json annotation does not allow any space inside ``.
// it will through warning and the annotation will not work
type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"` // Use the '-' directive to hide some this field from the response
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"` // the omitempty will hide the field only if the value is empty
	RunTime   Runtime     `json:"run_time,omitempty"` // if we add the string directive, the RunTime field will be shown as a string in the response
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}
