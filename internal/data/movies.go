package data

import (
	"time"

	"github.com/mostafejur21/greenlight_go/internal/validator"
)
const must_provided string = "must be provided"

// if we does not include the json annotation for the struct, the default struct value will
// the json keys. like "ID", "CreatedAt" etc.
// but if we include the struct annotation, then that will be the json keys.
// like ID int64 `json:"id"`
// Also note that the json annotation does not allow any space inside “.
// it will through warning and the annotation will not work
type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"` // Use the '-' directive to hide some this field from the response
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`     // the omitempty will hide the field only if the value is empty
	RunTime   Runtime   `json:"run_time,omitempty"` // if we add the string directive, the RunTime field will be shown as a string in the response
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	// Use the Check() method to execute our validation checks. this will add
	// provided key and error message to the errors map if the check does not evaluate
	// to true.
	v.Check(movie.Title != "", "title", must_provided)
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", must_provided)
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.RunTime != 0, "runtime", must_provided)
	v.Check(movie.RunTime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", must_provided)
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genres")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")

	// Note: using the validators Unique() method to check weather the Genres has unique slice or not
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}
