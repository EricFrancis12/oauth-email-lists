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

	router.HandleFunc("/outputs", handleInsertNewOutput).Methods(http.MethodPost)
	router.HandleFunc("/outputs", handleGetAllOutputs).Methods(http.MethodGet)
	router.HandleFunc("/outputs/{outputID}", handleGetOutputByID).Methods(http.MethodGet)

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

	// oauthID can be decoded to get the emailListID, providerName, and outputNames
	emailListID, provider, outputIDs, err := decenc.Decode(oauthID)
	if err != nil {
		RedirectToCatchAllUrl(w, r)
		return
	}

	NewProviderCookie(emailListID, provider.Name(), outputIDs).Set(w)

	provider.Redirect(w, r)
}

func handleGoogleOAuth(w http.ResponseWriter, r *http.Request) {
	emailListID := mux.Vars(r)["emailListID"]
	if emailListID == "" {
		RedirectToCatchAllUrl(w, r)
		return
	}

	provider := NewOAuthProvider(ProviderNameGoogle)
	NewProviderCookie(emailListID, provider.Name(), []string{}).Set(w)
	provider.Redirect(w, r)
}

var googleOAuthStateString = uuid.NewString()

func handleGoogleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, struct{}{})

	state := r.URL.Query().Get("state")
	if state != googleOAuthStateString {
		return
	}

	googleOauthConfig := config.Google()

	code := r.URL.Query().Get("code")
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return
	}

	client := googleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var gpr GoogleProviderResult
	if err := json.NewDecoder(bytes.NewReader(b)).Decode(&gpr); err != nil {
		return
	}

	pc, err := ProviderCookieFrom(r)
	if err != nil {
		return
	}

	go func() {
		for _, outputID := range pc.OutputIDs {
			output, err := storage.GetOutputByID(outputID)
			if err != nil {
				continue
			}
			output.Handle(gpr.Email, gpr.Name)
		}
	}()

	cr := SubscriberCreationReq{
		EmailListID: pc.EmailListID,
		Name:        gpr.Name,
		EmailAddr:   gpr.Email,
	}

	storage.InsertNewSubscriber(cr)
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

func handleInsertNewOutput(w http.ResponseWriter, r *http.Request) {
	var cr OutputCreationReq
	err := json.NewDecoder(r.Body).Decode(&cr)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}

	output, err := storage.InsertNewOutput(cr)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}

	WriteJSON(w, http.StatusOK, output)
}

func handleGetAllOutputs(w http.ResponseWriter, r *http.Request) {
	outputs, err := storage.GetAllOutputs()
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, NewServerErrorFromStr("output ID not provided"))
		return
	}
	WriteJSON(w, http.StatusOK, outputs)
}

func handleGetOutputByID(w http.ResponseWriter, r *http.Request) {
	outputID := mux.Vars(r)["outputID"]
	if outputID == "" {
		WriteJSON(w, http.StatusBadRequest, NewServerErrorFromStr("output ID not provided"))
		return
	}

	output, err := storage.GetOutputByID(outputID)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, NewServerErrorFromStr("output ID not provided"))
		return
	}

	WriteJSON(w, http.StatusOK, output)
}
