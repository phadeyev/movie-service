package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/dmitrii.fadeev/geek/schema"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
)

const Port = 8081

var db *sqlx.DB

func main() {
	var err error
	dbConnString := getEnvValue("MOVIE_DB_CONN_STR", "user=movie password=movie dbname=movie sslmode=disable host=localhost port=5432")

	db, err = sqlx.Open("postgres", dbConnString)
	if err != nil {
		panic(err.Error())
	}

	flag.Parse()
	switch flag.Arg(0) {
	case "migrate":
		if err := schema.Migrate(db); err != nil {
			log.Fatal("applying migrations", err)
		}
		log.Println("Migrations complete")
		return

	case "seed":
		if err := schema.Seed(db); err != nil {
			log.Fatal("applying seed data", err)
		}
		log.Println("Seed data inserted")
		return
	}
	r := mux.NewRouter()
	r.HandleFunc("/movies", movieListHandler)
	r.HandleFunc("/movie/{id}", movieByIdHanlder)
	log.Printf("Starting on port %d", Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(Port), r))
}

type Movie struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Director       string    `json:"director"`
	Year           int       `json:"year"`
	AgeRating      int       `json:"age_rating"`
	Poster         string    `json:"poster"`
	YoutubeVideoId string    `json:"movie_url"`
	IsPaid         bool      `json:"is_paid"`
}

func movieListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	movies, err := getMoviesFromDB()
	if err != nil {
		log.Printf("Can't get movie list form DB: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	err = json.NewEncoder(w).Encode(movies)
	if err != nil {
		log.Printf("Render response error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	return
}

func movieByIdHanlder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	vars := mux.Vars(r)
	id := vars["id"]
	uuid, err := uuid.FromString(id)
	if err != nil {
		log.Printf("Wrong movie ID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	movie, err := getMoviewById(uuid)
	if err != nil {
		log.Printf("Can't get movie by ID DB: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	err = json.NewEncoder(w).Encode(movie)
	if err != nil {
		log.Printf("Render response error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	return
}

func getEnvValue(env string, def string) (value string) {
	value, ok := os.LookupEnv(env)
	if ok != true {
		return def
	}
	return value
}

func getMoviesFromDB() ([]Movie, error) {
	rows, err := db.Query("SELECT movie_id, name, description, director, year, ageRating, poster, youtubeVideoId, isPaid from movies")
	movies := make([]Movie, 0)
	var movie Movie
	for rows.Next() {
		err := rows.Scan(&movie.ID, &movie.Name, &movie.Description, &movie.Director, &movie.Year, &movie.AgeRating, &movie.Poster, &movie.YoutubeVideoId, &movie.IsPaid)
		if err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}
	if err != nil {
		return nil, err
	}
	return movies, nil
}

func getMoviewById(uuid uuid.UUID) (Movie, error) {
	var movie Movie
	row := db.QueryRow("SELECT movie_id, name, description, director, year, ageRating, poster, youtubeVideoId, isPaid from movies where movie_id=$1", uuid.String())
	err := row.Scan(&movie.ID, &movie.Name, &movie.Description, &movie.Director, &movie.Year, &movie.AgeRating, &movie.Poster, &movie.YoutubeVideoId, &movie.IsPaid)
	if err != nil {
		if err == sql.ErrNoRows {
			return movie, fmt.Errorf("Moview not found by id")
		} else {
			return movie, err
		}
	}
	return movie, nil
}
