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

// TODO: replace all json responses with ServerResponseJSON
type ServerResponseJSON struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewServerResponseJSON(success bool, data any, err error) *ServerResponseJSON {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	return &ServerResponseJSON{
		Success: success,
		Data:    data,
		Error:   errMsg,
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

// TODO: add auth to server
func (s *Server) Run() error {
	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	router.HandleFunc("/c", Auth(handleMakeCampaign)).Methods(http.MethodPost)
	router.HandleFunc("/c", handleCampaign).Methods(http.MethodGet)

	router.HandleFunc("/t/google/{emailListID}", handleGoogleCampaign).Methods(http.MethodGet)
	router.HandleFunc("/callback/google", handleGoogleCampaignCallback).Methods(http.MethodGet)

	router.HandleFunc("/users", RootAuth(handleInsertNewUser)).Methods(http.MethodPost)
	router.HandleFunc("/users", RootAuth(handleGetAllUsers)).Methods(http.MethodGet)
	router.HandleFunc("/users/{userID}", RootAuth(handleGetUserByID)).Methods(http.MethodGet)
	router.HandleFunc("/users/{userID}", RootAuth(handleUpdateUserByID)).Methods(http.MethodPatch)
	router.HandleFunc("/users/{userID}", RootAuth(handleDeleteUserByID)).Methods(http.MethodDelete)

	router.HandleFunc("/email-lists", handleInsertNewEmailListByUserID).Methods(http.MethodPost)
	router.HandleFunc("/email-lists", handleGetAllEmailListsByUserID).Methods(http.MethodGet)

	router.HandleFunc("/subscribers", handleGetAllSubscribersByUserID).Methods(http.MethodGet)

	router.HandleFunc("/outputs", handleInsertNewOutputByUserID).Methods(http.MethodPost)
	router.HandleFunc("/outputs", handleGetAllOutputsByUserID).Methods(http.MethodGet)
	router.HandleFunc("/outputs/{outputID}", handleGetOutputByIDAndUserID).Methods(http.MethodGet)

	listenAddr := FallbackIfEmpty(s.ListenAddr, defaultListenAddr)
	fmt.Printf("Server running at %s\n", listenAddr)
	return http.ListenAndServe(listenAddr, router)
}

func handleCampaign(w http.ResponseWriter, r *http.Request) {
	oauthID := r.URL.Query().Get(QueryParamC)

	if oauthID == "" {
		RedirectToCatchAllUrl(w, r)
		return
	}

	// oauthID can be decoded to get the emailListID, providerName, and outputIDs
	emailListID, provider, outputIDs, err := decenc.Decode(oauthID)
	if err != nil {
		RedirectToCatchAllUrl(w, r)
		return
	}

	NewProviderCookie(emailListID, provider.Name(), outputIDs).Set(w)

	provider.Redirect(w, r)
}

func handleGoogleCampaign(w http.ResponseWriter, r *http.Request) {
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

func handleGoogleCampaignCallback(w http.ResponseWriter, r *http.Request) {
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

	emailList, err := storage.GetEmailListByID(pc.EmailListID)
	if err != nil {
		return
	}
	userID := emailList.UserID

	go func() {
		for _, outputID := range pc.OutputIDs {
			output, err := storage.GetOutputByIDAndUserID(outputID, userID)
			if err != nil {
				continue
			}
			output.Handle(gpr.Email, gpr.Name)
		}
	}()

	cr := SubscriberCreationReq{
		EmailListID: pc.EmailListID,
		UserID:      userID,
		Name:        gpr.Name,
		EmailAddr:   gpr.Email,
	}

	storage.InsertNewSubscriber(cr)
}

func handleMakeCampaign(w http.ResponseWriter, r *http.Request) {
	var c Campaign
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}

	link, err := c.Link()
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}

	WriteJSON(w, http.StatusOK, NewServerResponseJSON(true, link, nil))
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

func handleInsertNewEmailListByUserID(w http.ResponseWriter, r *http.Request) {
	user, err := useProtectedRoute(w, r)
	if err != nil {
		WriteUnauthorized(w)
		return
	}

	var cr EmailListCreationReq
	if err := json.NewDecoder(r.Body).Decode(&cr); err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}
	if !IsRootUser(user) {
		cr.UserID = user.ID
	}

	emailList, err := storage.InsertNewEmailList(cr)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}

	WriteJSON(w, http.StatusOK, emailList)
}

func handleGetAllEmailListsByUserID(w http.ResponseWriter, r *http.Request) {
	var (
		emailLists []*EmailList
		err        error
	)

	user, err := useProtectedRoute(w, r)
	if err != nil {
		WriteUnauthorized(w)
		return
	}

	if IsRootUser(user) {
		emailLists, err = storage.GetAllEmailLists()
	} else {
		emailLists, err = storage.GetAllEmailListsByUserID(user.ID)
	}
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}

	WriteJSON(w, http.StatusOK, emailLists)
}

func handleGetAllSubscribersByUserID(w http.ResponseWriter, r *http.Request) {
	var (
		subscribers []*Subscriber
		err         error
	)

	user, err := useProtectedRoute(w, r)
	if err != nil {
		WriteUnauthorized(w)
		return
	}

	if IsRootUser(user) {
		subscribers, err = storage.GetAllSubscribers()
	} else {
		subscribers, err = storage.GetAllSubscribersByUserID(user.ID)
	}
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}

	WriteJSON(w, http.StatusOK, subscribers)
}

func handleInsertNewOutputByUserID(w http.ResponseWriter, r *http.Request) {
	user, err := useProtectedRoute(w, r)
	if err != nil {
		WriteUnauthorized(w)
		return
	}

	var cr OutputCreationReq
	if err := json.NewDecoder(r.Body).Decode(&cr); err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}
	if !IsRootUser(user) {
		cr.UserID = user.ID
	}

	output, err := storage.InsertNewOutput(cr)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewServerError(err))
		return
	}

	WriteJSON(w, http.StatusOK, output)
}

func handleGetAllOutputsByUserID(w http.ResponseWriter, r *http.Request) {
	var (
		outputs []Output
		err     error
	)

	user, err := useProtectedRoute(w, r)
	if err != nil {
		WriteUnauthorized(w)
		return
	}

	if IsRootUser(user) {
		outputs, err = storage.GetAllOutputs()
	} else {
		outputs, err = storage.GetAllOutputsByUserID(user.ID)
	}
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, NewServerErrorFromStr("output ID not provided"))
		return
	}
	WriteJSON(w, http.StatusOK, outputs)
}

func handleGetOutputByIDAndUserID(w http.ResponseWriter, r *http.Request) {
	var (
		output Output
		err    error
	)

	user, err := useProtectedRoute(w, r)
	if err != nil {
		WriteUnauthorized(w)
		return
	}

	outputID := mux.Vars(r)["outputID"]
	if outputID == "" {
		WriteJSON(w, http.StatusBadRequest, NewServerErrorFromStr("output ID not provided"))
		return
	}

	if IsRootUser(user) {
		output, err = storage.GetOutputByID(outputID)
	} else {
		output, err = storage.GetOutputByIDAndUserID(outputID, user.ID)
	}
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, NewServerErrorFromStr("output ID not provided"))
		return
	}

	WriteJSON(w, http.StatusOK, output)
}
