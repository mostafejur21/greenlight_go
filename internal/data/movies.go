package data

import (
	"context"
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Create an args slice containing the values for the placeholder params from the movies struct.
	args := []any{movie.Title, movie.Year, movie.RunTime, pq.Array(movie.Genres)}

	// Use the DB.QueryRow() method to execute the SQL query on our connection pool,
	// passing in the args slice as a variadic parameters and scanning the system-generated id, created_at and version value into the movies struct
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
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

	// Use the context.WithTimeout() function to create a context.Context which carries a
	// 3-second timeout deadline. we are using a empty context.Background() as the 'parent' context.
	// the reason for using context is it will wait 3 second for executing this Get() method
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// Importantly, use defer to make sure that we cancel the context before the Get() method return.
	// the defer simple put it at the very end of the function line,
	defer cancel()

	// execute the query using the QueryRowContext() {new add the context} method, passing in the provided id value
	// as a placeholder parameters and scan the response data into the fields of the Movie struct.
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
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

// GetAll() method which will returns a slice of movies.
func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, error) {
	query := `
    SELECT id, created_at, title, year, runtime, genres, version
    FROM movies
    WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
    AND (genres @> $2 OR $2 = '{}')
    ORDER BY id`

	// Create a context with a 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use the QueryContext() to execute the query. This will return a sql.Rows result set
	// containing the result
	rows, err := m.DB.QueryContext(ctx, query, title, pq.Array(genres))
	if err != nil {
		return nil, err
	}

	// defer the rows to ensure that the result set is closed before GetAll() returns.
	defer rows.Close()

	// initialize an empty slice to hold the movie data.
	movies := []*Movie{}

	// Use rows.Next to iterate through the rows in the resultset.
	for rows.Next() {
		var movie Movie

		err := rows.Scan(
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.RunTime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, err
		}

		movies = append(movies, &movie)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return movies, nil
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Use the QueryRow() method to execute the query, passing in the args slice as
	// variadic parameter and scanning the new version value into the movie struct
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// execute the SQL query using the Exec() method, passing the id variable as the value for the
	// placeholder parameters, The Exec() method return a sql.Result object
	result, err := m.DB.ExecContext(ctx, query, id)
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
