package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const defaultListenAddr string = ":6004"

type Server struct {
	ListenAddr string
}

func NewServer(listenAddr string) *Server {
	return &Server{
		ListenAddr: listenAddr,
	}
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

func (a *Server) Run() error {
	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	router.HandleFunc("/g/{oauthID}", handleOAuth).Methods(http.MethodGet)
	router.HandleFunc("/t/google/{emailListID}", handleGoogleOAuth).Methods(http.MethodGet)
	router.HandleFunc("/callback/google", handleGoogleOAuthCallback).Methods(http.MethodGet)

	router.HandleFunc("/users", handleInsertNewUser).Methods(http.MethodPost)
	router.HandleFunc("/users", handleGetAllUsers).Methods(http.MethodGet)
	router.HandleFunc("/users/{userID}", handleGetUserByID).Methods(http.MethodGet)
	router.HandleFunc("/users/{userID}", handleUpdateUserByID).Methods(http.MethodPatch)
	router.HandleFunc("/users/{userID}", handleDeleteUserByID).Methods(http.MethodDelete)

	router.HandleFunc("/email-lists", handleInsertNewEmailList).Methods(http.MethodPost)
	router.HandleFunc("/email-lists", handleGetAllEmailLists).Methods(http.MethodGet)

	router.HandleFunc("/subscribers", handleGetAllSubscribers).Methods(http.MethodGet)

	listenAddr := FallbackIfEmpty(a.ListenAddr, defaultListenAddr)
	fmt.Printf("Server running at %s\n", listenAddr)
	return http.ListenAndServe(listenAddr, router)
}

func handleOAuth(w http.ResponseWriter, r *http.Request) {
	oauthID := mux.Vars(r)["oauthID"]
	if oauthID == "" {
		RedirectToCatchAllUrl(w, r)
		return
	}

	// oauthID can be decoded to get the emailListID and providerName
	emailListID, provider, err := decenc.Decode(oauthID)
	if err != nil {
		RedirectToCatchAllUrl(w, r)
		return
	}

	NewProviderCookie(emailListID, provider.Name()).Set(w)

	provider.Redirect(w, r)
}

func handleGoogleOAuth(w http.ResponseWriter, r *http.Request) {
	emailListID := mux.Vars(r)["emailListID"]
	if emailListID == "" {
		RedirectToCatchAllUrl(w, r)
		return
	}

	provider := NewOAuthProvider(ProviderNameGoogle)
	NewProviderCookie(emailListID, provider.Name()).Set(w)
	provider.Redirect(w, r)
}

var googleOAuthStateString = uuid.NewString()

func handleGoogleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	fmt.Println("~ 0")

	WriteJSON(w, http.StatusOK, struct{}{})

	fmt.Println("~ 1")

	state := r.URL.Query().Get("state")
	if state != googleOAuthStateString {
		return
	}

	fmt.Println("~ 2")

	googleOauthConfig := config.Google()

	code := r.URL.Query().Get("code")
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return
	}

	fmt.Println("~ 3")

	client := googleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	fmt.Println("~ 4")

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	fmt.Println("~ 5")

	var gpr GoogleProviderResult
	if err := json.NewDecoder(bytes.NewReader(b)).Decode(&gpr); err != nil {
		return
	}

	fmt.Println("~ 6")

	pc, err := ProviderCookieFrom(r)
	if err != nil {
		return
	}

	fmt.Println("~ 7")

	cr := SubscriberCreationReq{
		EmailListID: pc.EmailListID,
		Name:        gpr.Name,
		EmailAddr:   gpr.Email,
	}
	if _, err := storage.InsertNewSubscriber(cr); err != nil {
		fmt.Println("~ 8")
		fmt.Println(err.Error())
		return
	}

	fmt.Println("~ 9")
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

func handleGetAllUsers(w http.ResponseWriter, _ *http.Request) {
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

func handleInsertNewEmailList(w http.ResponseWriter, r *http.Request) {
	var cr EmailListCreationReq
	err := json.NewDecoder(r.Body).Decode(&cr)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}

	emailList, err := storage.InsertNewEmailList(cr)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}

	WriteJSON(w, http.StatusOK, emailList)
}

func handleGetAllEmailLists(w http.ResponseWriter, _ *http.Request) {
	emailLists, err := storage.GetAllEmailLists()
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}

	WriteJSON(w, http.StatusOK, emailLists)
}

func handleGetAllSubscribers(w http.ResponseWriter, _ *http.Request) {
	subscribers, err := storage.GetAllSubscribers()
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}

	WriteJSON(w, http.StatusOK, subscribers)
}
