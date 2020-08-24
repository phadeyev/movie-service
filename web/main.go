package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/dmitrii.fadeev/geek/pkg/render"
	"github.com/dmitrii.fadeev/geek/pkg/requester"
	uuid "github.com/satori/go.uuid"

	"github.com/gorilla/mux"
)

type Config struct {
	Port        int
	UserAddr    string
	MovieAddr   string
	PaymentAddr string
}

var cfg Config

func main() {
	cfg.Port = getEnvValueInt("PORT", 8080)
	cfg.MovieAddr = getEnvValue("MOVIE_ADDR", "http://movie:8081")
	cfg.UserAddr = getEnvValue("USER_ADDR", "http://movie:8082")
	cfg.PaymentAddr = getEnvValue("PAYMENT_ADDR", "http://movie:8082")

	r := mux.NewRouter()
	r.HandleFunc("/", MainHandler)
	r.HandleFunc("/movie/{id}", MovieHandler)

	// Обработчик статических файлов
	fs := http.FileServer(http.Dir("assets"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Настройка шаблонизатора
	render.SetTemplateDir(".")
	render.SetTemplateLayout("layout.html")
	render.AddTemplate("main", "main.html")
	render.AddTemplate("login", "login.html")
	render.AddTemplate("movie", "movie.html")
	err := render.ParseTemplates()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting on port %d", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(cfg.Port), r))
}

type MainPage struct {
	Movies *[]Movie
	User   User
	PayURL string
}

type MoviePage struct {
	Movie  Movie
	User   User
	PayURL string
}

type User struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	IsPaid bool   `json:"is_paid"`
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

func MainHandler(w http.ResponseWriter, r *http.Request) {
	page := MainPage{}

	var err error
	page.Movies, err = getMovies()
	if err != nil {
		log.Printf("Get movie error: %v", err)
	}

	page.User, err = getUser(r)
	if err != nil {
		log.Printf("Get user error: %v", err)
	} else {
		page.PayURL = cfg.PaymentAddr + "/checkout?uid=" + strconv.Itoa(page.User.ID)
	}

	render.RenderTemplate(w, "main", page)
}

func MovieHandler(w http.ResponseWriter, r *http.Request) {
	page := MoviePage{}
	vars := mux.Vars(r)
	id := vars["id"]
	uuid, err := uuid.FromString(id)
	if err != nil {
		log.Printf("Wrong movie ID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	movie, err := getMovie(uuid)
	if err != nil {
		log.Printf("Get movie error: %v", err)
	}
	page.Movie = *movie
	page.User, err = getUser(r)
	if err != nil {
		log.Printf("Get user error: %v", err)
	} else {
		page.PayURL = cfg.PaymentAddr + "/checkout?uid=" + strconv.Itoa(page.User.ID)
	}

	render.RenderTemplate(w, "movie", page)
}

func getMovies() (*[]Movie, error) {
	mm := &[]Movie{}
	err := requester.GetJSON(cfg.MovieAddr+"/movies", mm)
	if err != nil {
		return nil, err
	}

	return mm, nil
}

func getMovie(uuid uuid.UUID) (*Movie, error) {
	m := &Movie{}
	err := requester.GetJSON(cfg.MovieAddr+"/movie/"+uuid.String(), m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func getUser(r *http.Request) (usr User, err error) {
	ses, err := r.Cookie("session")
	if ses == nil {
		return usr, err
	}

	res := &struct {
		User
		Error string
	}{}
	err = requester.GetJSON(cfg.UserAddr+"/user?token="+ses.Value, res)
	if err != nil {
		return usr, err
	}

	if res.Error != "" {
		return usr, fmt.Errorf(res.Error)
	}

	usr.ID = res.ID
	usr.Name = res.Name
	usr.IsPaid = res.IsPaid

	return usr, nil
}

func getEnvValue(env string, def string) (value string) {
	value, ok := os.LookupEnv(env)
	if ok != true {
		return def
	}
	return value
}

func getEnvValueInt(env string, def int) (value int) {
	valueStr := getEnvValue(env, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return def
}
