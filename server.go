package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

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

type JsonResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewJsonResponse(success bool, data any, err error) *JsonResponse {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	return &JsonResponse{
		Success: success,
		Data:    data,
		Error:   errMsg,
	}
}

func (s *Server) Run() error {
	router := mux.NewRouter()

	// Auth
	router.HandleFunc("/login", handleGetLogin).Methods(http.MethodGet)
	router.HandleFunc("/login", handlePostLogin).Methods(http.MethodPost)
	router.HandleFunc("/logout", handleLogout).Methods(http.MethodGet, http.MethodPost)

	// General campaigns
	router.HandleFunc("/c", Auth(handleMakeCampaign)).Methods(http.MethodPost)
	router.HandleFunc("/c", handleCampaign).Methods(http.MethodGet)

	// Discord campaigns
	router.HandleFunc("/t/discord/{emailListID}", handleDiscordCampaign).Methods(http.MethodGet)
	router.HandleFunc("/callback/discord", handleDiscordCampaignCallback).Methods(http.MethodGet)

	// Google campaigns
	router.HandleFunc("/t/google/{emailListID}", handleGoogleCampaign).Methods(http.MethodGet)
	router.HandleFunc("/callback/google", handleGoogleCampaignCallback).Methods(http.MethodGet)

	// Users
	router.HandleFunc("/users", RootAuth(handleInsertNewUser)).Methods(http.MethodPost)
	router.HandleFunc("/users", RootAuth(handleGetAllUsers)).Methods(http.MethodGet)
	router.HandleFunc("/users/{userID}", RootAuth(handleGetUserByID)).Methods(http.MethodGet)
	router.HandleFunc("/users/{userID}", RootAuth(handleUpdateUserByID)).Methods(http.MethodPatch)
	router.HandleFunc("/users/{userID}", RootAuth(handleDeleteUserByID)).Methods(http.MethodDelete)

	// Email lists
	router.HandleFunc("/email-lists", handleInsertNewEmailListByUserID).Methods(http.MethodPost)
	router.HandleFunc("/email-lists", handleGetAllEmailListsByUserID).Methods(http.MethodGet)

	// Subscribers
	router.HandleFunc("/subscribers", handleInsertNewSubscriberByEmailListIDAndUserID).Methods(http.MethodPost)
	router.HandleFunc("/subscribers", handleGetAllSubscribersByUserID).Methods(http.MethodGet)

	// Outputs
	router.HandleFunc("/outputs", handleInsertNewOutputByUserID).Methods(http.MethodPost)
	router.HandleFunc("/outputs", handleGetAllOutputsByUserID).Methods(http.MethodGet)
	router.HandleFunc("/outputs/{outputID}", handleGetOutputByIDAndUserID).Methods(http.MethodGet)
	router.HandleFunc("/outputs/{outputID}", handleUpdateOutputByIDAndUserID).Methods(http.MethodPatch)

	// Misc
	router.HandleFunc("/healthz", handleHealthz)
	router.HandleFunc("/", handleCatchAll)
	router.HandleFunc(`/{catchAll:[a-zA-Z0-9=\-\/.]+}`, handleCatchAll)

	listenAddr := FallbackIfEmpty(s.ListenAddr, defaultListenAddr)
	fmt.Printf("Server running at %s\n", listenAddr)
	return http.ListenAndServe(listenAddr, router)
}

func handleGetLogin(w http.ResponseWriter, r *http.Request) {
	b, err := os.ReadFile(filePathLoginPage)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}
	w.Write(b)
}

func handlePostLogin(w http.ResponseWriter, r *http.Request) {
	var li LoginInfo
	err := json.NewDecoder(r.Body).Decode(&li)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}

	user, err := storage.GetUserByUsernameAndPassword(li.Username, li.Password)
	if err != nil || user == nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}

	if err := Login(w, user); err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}

	WriteJSON(w, http.StatusOK, NewJsonResponse(true, nil, nil))
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	Logout(w)
	WriteJSON(w, http.StatusOK, NewJsonResponse(true, nil, nil))
}

func handleCampaign(w http.ResponseWriter, r *http.Request) {
	oauthID := r.URL.Query().Get(QueryParamC)

	if oauthID == "" {
		RedirectToCatchAllUrl(w, r)
		return
	}

	// oauthID can be decoded to get the emailListID, providerName, and outputIDs
	emailListID, provider, outputIDs, redirectUrl, err := decenc.Decode(oauthID)
	if err != nil {
		RedirectToCatchAllUrl(w, r)
		return
	}

	pc := NewProviderCookie(emailListID, provider.Name(), outputIDs, redirectUrl)
	if err := pc.Set(w); err != nil {
		RedirectToCatchAllUrl(w, r)
		return
	}

	provider.Redirect(w, r)
}

func handleDiscordCampaign(w http.ResponseWriter, r *http.Request) {
	// TODO: ...
}

func handleDiscordCampaignCallback(w http.ResponseWriter, r *http.Request) {
	// TODO: ...
}

func handleGoogleCampaign(w http.ResponseWriter, r *http.Request) {
	var (
		outputIDs   = r.URL.Query()["o"]
		redirectUrl = r.URL.Query().Get("r")
	)

	emailListID := mux.Vars(r)["emailListID"]
	if emailListID == "" {
		RedirectToCatchAllUrl(w, r)
		return
	}

	provider := NewOAuthProvider(ProviderNameGoogle)

	pc := NewProviderCookie(emailListID, provider.Name(), outputIDs, redirectUrl)
	if err := pc.Set(w); err != nil {
		RedirectToCatchAllUrl(w, r)
		return
	}

	provider.Redirect(w, r)
}

var googleOAuthStateString = uuid.NewString()

func handleGoogleCampaignCallback(w http.ResponseWriter, r *http.Request) {
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

	go RedirectVisitor(w, r, pc.RedirectUrl)

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
		EmailListID:        pc.EmailListID,
		UserID:             userID,
		SourceProviderName: pc.ProviderName,
		Name:               gpr.Name,
		EmailAddr:          gpr.Email,
	}
	storage.InsertNewSubscriber(cr)
}

func handleMakeCampaign(w http.ResponseWriter, r *http.Request) {
	var c Campaign
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}

	link, err := c.Link()
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}

	WriteJSON(w, http.StatusOK, NewJsonResponse(true, link, nil))
}

func handleInsertNewUser(w http.ResponseWriter, r *http.Request) {
	var cr UserCreationReq
	err := json.NewDecoder(r.Body).Decode(&cr)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}

	user, err := storage.InsertNewUser(cr)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}

	WriteJSON(w, http.StatusOK, NewJsonResponse(true, user, nil))
}

func handleGetAllUsers(w http.ResponseWriter, _ *http.Request) {
	users, err := storage.GetAllUsers()
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}
	WriteJSON(w, http.StatusOK, NewJsonResponse(true, users, nil))
}

func handleGetUserByID(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	if userID == "" {
		WriteJSON(w, http.StatusBadRequest, NewJsonResponse(false, nil, userIDNotProvided()))
		return
	}

	user, err := storage.GetUserByID(userID)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}

	WriteJSON(w, http.StatusOK, NewJsonResponse(true, user, nil))
}

func handleUpdateUserByID(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	if userID == "" {
		WriteJSON(w, http.StatusBadRequest, NewJsonResponse(false, nil, userIDNotProvided()))
		return
	}

	var ur UserUpdateReq
	err := json.NewDecoder(r.Body).Decode(&ur)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}

	if err := storage.UpdateUserByID(userID, ur); err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}

	WriteJSON(w, http.StatusOK, NewJsonResponse(true, nil, nil))
}

func handleDeleteUserByID(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	if userID == "" {
		WriteJSON(w, http.StatusBadRequest, NewJsonResponse(false, nil, userIDNotProvided()))
		return
	}

	if err := storage.DeleteUserByID(userID); err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}

	WriteJSON(w, http.StatusOK, NewJsonResponse(true, nil, nil))
}

func handleInsertNewEmailListByUserID(w http.ResponseWriter, r *http.Request) {
	user, err := useProtectedRoute(w, r)
	if err != nil {
		WriteUnauthorized(w)
		return
	}

	var cr EmailListCreationReq
	if err := json.NewDecoder(r.Body).Decode(&cr); err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}
	if !IsRootUser(user) {
		cr.UserID = user.ID
	}

	emailList, err := storage.InsertNewEmailList(cr)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}

	WriteJSON(w, http.StatusOK, NewJsonResponse(true, emailList, nil))
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
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}

	WriteJSON(w, http.StatusOK, NewJsonResponse(true, emailLists, nil))
}

func handleInsertNewSubscriberByEmailListIDAndUserID(w http.ResponseWriter, r *http.Request) {
	user, err := useProtectedRoute(w, r)
	if err != nil {
		WriteUnauthorized(w)
		return
	}

	var cr SubscriberCreationReq
	if err := json.NewDecoder(r.Body).Decode(&cr); err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}
	if !IsRootUser(user) {
		cr.UserID = user.ID
	}

	subscriber, err := storage.InsertNewSubscriber(cr)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}

	WriteJSON(w, http.StatusOK, NewJsonResponse(true, subscriber, nil))
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
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}

	WriteJSON(w, http.StatusOK, NewJsonResponse(true, subscribers, nil))
}

func handleInsertNewOutputByUserID(w http.ResponseWriter, r *http.Request) {
	user, err := useProtectedRoute(w, r)
	if err != nil {
		WriteUnauthorized(w)
		return
	}

	var cr OutputCreationReq
	if err := json.NewDecoder(r.Body).Decode(&cr); err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}
	if !IsRootUser(user) {
		cr.UserID = user.ID
	}

	output, err := storage.InsertNewOutput(cr)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}

	WriteJSON(w, http.StatusOK, NewJsonResponse(true, output, nil))
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
		WriteJSON(w, http.StatusBadRequest, NewJsonResponse(false, nil, outputIDNotProvided()))
		return
	}
	WriteJSON(w, http.StatusOK, NewJsonResponse(true, makeOutputsData(outputs), nil))
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
		WriteJSON(w, http.StatusBadRequest, NewJsonResponse(false, nil, outputIDNotProvided()))
		return
	}

	if IsRootUser(user) {
		output, err = storage.GetOutputByID(outputID)
	} else {
		output, err = storage.GetOutputByIDAndUserID(outputID, user.ID)
	}
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, NewJsonResponse(false, nil, outputIDNotProvided()))
		return
	}

	WriteJSON(w, http.StatusOK, output)
}

func handleUpdateOutputByIDAndUserID(w http.ResponseWriter, r *http.Request) {
	user, err := useProtectedRoute(w, r)
	if err != nil {
		WriteUnauthorized(w)
		return
	}

	outputID := mux.Vars(r)["outputID"]
	if outputID == "" {
		WriteJSON(w, http.StatusBadRequest, NewJsonResponse(false, nil, outputIDNotProvided()))
		return
	}

	var ur OutputUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&ur); err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}

	userID := user.ID
	if IsRootUser(user) {
		output, err := storage.GetOutputByID(outputID)
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
			return
		}
		userID = output.GetUserID()
	}

	if err := storage.UpdateOutputByIDAndUserID(outputID, userID, ur); err != nil {
		WriteJSON(w, http.StatusInternalServerError, NewJsonResponse(false, nil, err))
		return
	}

	WriteJSON(w, http.StatusOK, NewJsonResponse(true, nil, nil))
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, struct{}{})
}

func handleCatchAll(w http.ResponseWriter, r *http.Request) {
	RedirectToCatchAllUrl(w, r)
}
