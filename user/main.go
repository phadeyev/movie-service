package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/dmitrii.fadeev/geek/pkg/render"

	"github.com/gorilla/mux"
)

const Port = 8082

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/user", userHandler).Methods("GET")
	r.HandleFunc("/user", userPatchHandler).Methods("PATCH")
	log.Printf("Starting on port %d", Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(Port), r))
}

type User struct {
	ID     int    `json:"id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	IsPaid bool   `json:"is_paid"`
	Pwd    string `json:"-"`
	Token  string `json:"token"`
}

type UserStorage []*User

var UU = UserStorage{
	&User{1, "bob@mail.ru", "Bob", true, "god", "1"},
	&User{2, "alice@mail.ru", "Alice", false, "secret", "2"},
}

func (uu UserStorage) GetByToken(token string) *User {
	for _, u := range uu {
		if u.Token == token {
			return u
		}
	}
	return nil
}

func (uu UserStorage) GetByID(id int) *User {
	for _, u := range uu {
		if u.ID == id {
			return u
		}
	}
	return nil
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	token := r.Form.Get("token")

	usr := UU.GetByToken(token)

	if usr == nil {
		render.RenderJSONErr(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	render.RenderJSON(w, usr)
	return
}

func userPatchHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	idStr := r.FormValue("id")
	isPaidStr := r.FormValue("is_paid")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		render.RenderJSONErr(w, "Invalid 'id': "+err.Error(), http.StatusBadRequest)
		return
	}

	usr := UU.GetByID(id)

	isPaid, err := strconv.ParseBool(isPaidStr)
	if err != nil {
		render.RenderJSONErr(w, "Invalid 'is_paid': "+err.Error(), http.StatusBadRequest)
		return
	}

	if usr == nil {
		render.RenderJSONErr(w, "Пользователь не найден. ID: "+idStr, http.StatusNotFound)
		return
	}

	usr.IsPaid = isPaid

	render.RenderJSON(w, usr)
	return
}
