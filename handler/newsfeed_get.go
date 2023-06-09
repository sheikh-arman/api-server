package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/sheikh-arman/api-server/newsfeed"
)

var jwtkey = []byte("adsads")
var TokenAuth *jwtauth.JWTAuth
var tokenString string
var token jwt.Token

var ID int
var feeds []newsfeed.Item
var feeds2 map[int]newsfeed.Item
var Credslist map[string]string

func InitCred() {
	TokenAuth = jwtauth.New(string(jwa.HS256), jwtkey, nil)
	Credslist = make(map[string]string)

	creds := []newsfeed.Credentials{
		{
			Username: "arman",
			Password: "123",
		},
	}

	for _, cred := range creds {
		Credslist[cred.Username] = cred.Password
	}
}

func InitDB() {
	ID = 1
	var feed newsfeed.Item
	feed = newsfeed.Item{
		Id:    ID,
		Title: "Nothing",
		Post:  "Lorem Ipsum Doller Site",
	}
	//feeds2[ID] = feed
	ID++
	feeds = append(feeds, feed)

	feed = newsfeed.Item{
		Id:    ID,
		Title: "Nothing2",
		Post:  "Lorem Ipsum Doller Site2",
	}
	//feeds2[ID] = feed
	ID++
	feeds = append(feeds, feed)

	feed = newsfeed.Item{
		Id:    ID,
		Title: "Nothing3",
		Post:  "Lorem Ipsum Doller Site3",
	}
	//feeds2[ID] = feed
	ID++
	feeds = append(feeds, feed)
}

func WriteJsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func GetNewsFeeds(w http.ResponseWriter, r *http.Request) {
	log.Println("test")
	sort.SliceStable(feeds, func(i, j int) bool {
		return feeds[i].Id < feeds[j].Id
	})
	WriteJsonResponse(w, http.StatusOK, feeds2)
}

func GetNewsFeed(w http.ResponseWriter, r *http.Request) {
	param := chi.URLParam(r, "id")
	paramsID, _ := strconv.Atoi(param)

	for _, curFeed := range feeds {
		if curFeed.Id == paramsID {
			WriteJsonResponse(w, http.StatusOK, curFeed)
			return
		}
	}
	WriteJsonResponse(w, http.StatusNotFound, "Newsfeed doesn't exist")
}

func CreateNewsFeed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var newFeed newsfeed.Item
	err := json.NewDecoder(r.Body).Decode(&newFeed)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	newFeed.Id = ID
	feeds = append(feeds, newFeed)
	ID++
}
func DeleteNewsFeed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	param := chi.URLParam(r, "id")
	paramsID, _ := strconv.Atoi(param)
	for index, curFeed := range feeds {
		if curFeed.Id == paramsID {
			feeds = append(feeds[:index], feeds[index+1:]...)
			break
		}
	}
	//fmt.Println(feeds)
	WriteJsonResponse(w, http.StatusOK, feeds)
}
func UpdateNewsFeed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	param := chi.URLParam(r, "id")
	paramID, _ := strconv.Atoi(param)
	var newFeed newsfeed.Item
	err := json.NewDecoder(r.Body).Decode(&newFeed)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	for index, curFeed := range feeds {
		if curFeed.Id == paramID {
			newFeed.Id = paramID
			feeds[index] = newFeed
			json.NewEncoder(w).Encode(feeds)
			return
		}
	}
	json.NewEncoder(w).Encode("No data")
}
func Login(w http.ResponseWriter, r *http.Request) {
	var creds newsfeed.Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	fmt.Println(creds)

	if err != nil {
		fmt.Println("what1")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	correctPassword, ok := Credslist[creds.Username]

	if !ok || creds.Password != correctPassword {
		//fmt.Println("what2")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	expiretime := time.Now().Add(10 * time.Minute)

	_, tokenString, err := TokenAuth.Encode(map[string]interface{}{
		"aud": "arman",
		"exp": expiretime.Unix(),
	})
	if err != nil {
		//fmt.Println("what3")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Printf("token string = %s \n", tokenString)
	http.SetCookie(w, &http.Cookie{
		Name:    "jwt",
		Value:   tokenString,
		Expires: expiretime,
	})
	w.WriteHeader(http.StatusOK)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "jwt",
		Expires: time.Now(),
	})
	w.WriteHeader(http.StatusOK)
}
func StartServer(port int) {
	InitCred()
	InitDB()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Post("/login", Login)
	r.Group(func(r chi.Router) {
		// jwtauth-> will learn later
		r.Use(jwtauth.Verifier(TokenAuth))
		r.Use(jwtauth.Authenticator)
		r.Route("/newsfeeds", func(r chi.Router) {
			r.Get("/", GetNewsFeeds)
			r.Get("/{id}", GetNewsFeed)
			r.Post("/", CreateNewsFeed)
			r.Delete("/{id}", DeleteNewsFeed)
			r.Put("/{id}", UpdateNewsFeed)
		})
		r.Post("/logout", Logout)
	})
	//port := 5050
	Server := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: r,
	}
	fmt.Println("Serving on " + strconv.Itoa(port))
	//http.ListenAndServe(strconv.Itoa(port), r)
	fmt.Println(Server.ListenAndServe())
}
