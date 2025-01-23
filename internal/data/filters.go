package data

import "github.com/mostafejur21/greenlight_go/internal/validator"

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

func ValidateFilters(v *validator.Validator, f Filters) {
	// Check that the page and page_size parameters contain sensible values.
	// page 1 to 10,000,000
	// page_size 1 to 100

	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10000000, "page", "must be a maximum number of 10 million")

	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	//Check that the sort parameter match a value in the safelist
    v.Check(validator.PermittedValue(f.Sort, f.SortSafelist...), "sort", "invalit sort value")
}
