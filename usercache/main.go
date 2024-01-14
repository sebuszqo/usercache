package main

import (
	"encoding/json"
	"fmt"
	cachebucket "github.com/sebuszqo/usercache/cache_bucket"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type userResponse struct {
	ID    int    `json:"ID"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type errorResponse struct {
	Error string `json:"error"`
}

var uc *cachebucket.UserCache

func init() {
	uc = cachebucket.NewUserCache(50)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/user/{id:[0-9]+}", getUserHandler).Methods("GET")
	http.Handle("/", r)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(1)
	}
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr, ok := vars["id"]
	if !ok {
		handleError(w, fmt.Errorf("user ID not provided"))
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		handleError(w, fmt.Errorf("Invalid user ID"))
		return
	}

	user, err := uc.GetUser(uint(userID))
	if err != nil {
		handleError(w, err)
		return
	}

	userResp := userResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(userResp)
	if err != nil {
		handleError(w, err)
		return
	}
}

func handleError(w http.ResponseWriter, err error) {
	errorResp := errorResponse{
		Error: err.Error(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(errorResp)
}
