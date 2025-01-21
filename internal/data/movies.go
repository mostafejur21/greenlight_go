package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/mostafejur21/greenlight_go/internal/validator"
)

const must_provided string = "must be provided"

// Define a MovieModel struct type which wraps a sql.DB connection pool.
type MovieModel struct {
	DB *sql.DB
}

// Add a placeholder method for inserting a new record in the movies table
func (m MovieModel) Insert(movie *Movie) error {
	// Define a SQL query for inserting a new record in the movies table and returning the system-generated data
	query := `
        INSERT INTO movies (title, year, runtime, genres)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, version`

	// Create an args slice containing the values for the placeholder params from the movies struct.
	args := []any{movie.Title, movie.Year, movie.RunTime, pq.Array(movie.Genres)}

	// Use the DB.QueryRow() method to execute the SQL query on our connection pool,
	// passing in the args slice as a variadic parameters and scanning the system-generated id, created_at and version value into the movies struct
	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Add a placeholder method for getting/fetching record from movies table

func (m MovieModel) Get(id int64) (*Movie, error) {
	// the postgreSQL bigserial type start with auto-incremental at 1 by default.
	if id < 1 {
		return nil, ErrInvalidRunTimeFormat
	}

	// Define the sql query
	query := `
        SELECT id, created_at, title, year, runtime, genres, version
        FROM movies
        WHERE id = $1`
	var movie Movie

	// execute the query using the QueryRow() method, passing in the provided id value
	// as a placeholder parameters and scan the response data into the fields of the Movie struct.
	err := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.RunTime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	// Handle any errors. if there was no matching movie found, Scan() will
	// return a sql.ErrNoRows error. we check for this and return our custom error
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// otherwise return a pointer to the movies struct
	return &movie, nil

}

func (m MovieModel) Update(movie *Movie) error {
	// SQL query for updating method
	query := `
    UPDATE movies
    SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
    WHERE id = $5 AND version = $6
    RETURNING version`
	// Create an args slice containing the values for the placeholder parameters.
	args := []any{
		movie.Title,
		movie.Year,
		movie.RunTime,
		pq.Array(movie.Genres),
		movie.ID,
	}
	// Use the QueryRow() method to execute the query, passing in the args slice as
	// variadic parameter and scanning the new version value into the movie struct
	err := m.DB.QueryRow(query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m MovieModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the id is less then 1
	if id < 1 {
		return ErrRecordNotFound
	}

	// SQL query
	query := `
    DELETE FROM movies
    WHERE id = $1`

	// execute the SQL query using the Exec() method, passing the id variable as the value for the
	// placeholder parameters, The Exec() method return a sql.Result object
	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}
	// Calling the RowsAffected() method on the sql.Result object to get the number of
	// Rows affected by the query
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// If no rows were affected, we know that the movies table didnot contain a record
	// with the provided ID at the moment we tried to delete it. In that case, we return an ErrRecordNotFound error
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

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
