package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

const defaultListenAddr string = ":6004"

type Server struct {
	ListenAddr string
}

type ServerError struct {
	Error string `json:"error"`
}

func NewServerError(err error) ServerError {
	return ServerError{
		Error: err.Error(),
	}
}

func NewServerErrorFromStr(errMsg string) ServerError {
	return ServerError{
		Error: errMsg,
	}
}

func NewServer(listenAddr string) *Server {
	return &Server{
		ListenAddr: listenAddr,
	}
}

func (a *Server) Run() error {
	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	router.HandleFunc("/users", handleInsertNewUser).Methods(http.MethodPost)
	router.HandleFunc("/users", handleGetAllUsers).Methods(http.MethodGet)

	router.HandleFunc("/users/{userID}", handleGetUserByID).Methods(http.MethodGet)
	router.HandleFunc("/users/{userID}", handleUpdateUserByID).Methods(http.MethodPatch)
	router.HandleFunc("/users/{userID}", handleDeleteUserByID).Methods(http.MethodDelete)

	listenAddr := FallbackIfEmpty(a.ListenAddr, defaultListenAddr)
	fmt.Printf("Server running at %s\n", listenAddr)
	return http.ListenAndServe(listenAddr, router)
}

func handleInsertNewUser(w http.ResponseWriter, r *http.Request) {
	var cr UserCreationReq
	err := json.NewDecoder(r.Body).Decode(&cr)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}

	user, err := storage.InsertNewUser(cr)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}

	WriteJSON(w, http.StatusOK, user)
}

func handleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := storage.GetAllUsers()
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}

	WriteJSON(w, http.StatusOK, users)
}

func handleGetUserByID(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	if userID == "" {
		WriteJSON(w, http.StatusBadRequest, NewServerErrorFromStr("user ID not provided"))
		return
	}

	user, err := storage.GetUserByID(userID)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}

	WriteJSON(w, http.StatusOK, user)
}

func handleUpdateUserByID(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	if userID == "" {
		WriteJSON(w, http.StatusBadRequest, NewServerErrorFromStr("user ID not provided"))
		return
	}

	var ur UserUpdateReq
	err := json.NewDecoder(r.Body).Decode(&ur)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}

	if err := storage.UpdateUserByID(userID, ur); err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}

	WriteJSON(w, http.StatusOK, struct{}{})
}

func handleDeleteUserByID(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	if userID == "" {
		WriteJSON(w, http.StatusBadRequest, NewServerErrorFromStr("user ID not provided"))
		return
	}

	if err := storage.DeleteUserByID(userID); err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}

	WriteJSON(w, http.StatusOK, struct{}{})
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
