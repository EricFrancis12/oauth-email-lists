package main

import (
	"net/http"
	"os"
	"strings"
	"time"
)

type ProviderCookie struct {
	EmailListID  string
	ProviderName ProviderName
	OutputIDs    []string
	RedirectUrl  string
	CreatedAt    time.Time
}

func NewProviderCookie(emailListID string, providerName ProviderName, outputIDs []string, redirectUrl string) *ProviderCookie {
	return &ProviderCookie{
		EmailListID:  emailListID,
		ProviderName: providerName,
		OutputIDs:    outputIDs,
		RedirectUrl:  redirectUrl,
		CreatedAt:    time.Now(),
	}
}

func ProviderCookieFrom(r *http.Request) (*ProviderCookie, error) {
	emailListID, err := CookieNameEmailListID.DecryptFrom(r)
	if err != nil {
		return nil, err
	}

	s, err := CookieNameProviderName.DecryptFrom(r)
	if err != nil {
		return nil, err
	}
	providerName, err := ToProviderName(s)
	if err != nil {
		return nil, err
	}

	outputIDsStr, err := CookieNameOutputIDs.DecryptFrom(r)
	if err != nil {
		return nil, err
	}
	outputIDs := strings.Split(outputIDsStr, outputCookieDelim)

	redirectUrl, err := CookieNameRedirectURL.DecryptFrom(r)
	if err != nil {
		return nil, err
	}

	createdAtStr, err := CookieNameCreatedAt.DecryptFrom(r)
	if err != nil {
		return nil, err
	}
	createdAt, err := time.Parse(timestampLayout, createdAtStr)
	if err != nil {
		return nil, err
	}

	return &ProviderCookie{
		EmailListID:  emailListID,
		ProviderName: providerName,
		OutputIDs:    outputIDs,
		RedirectUrl:  redirectUrl,
		CreatedAt:    createdAt,
	}, nil
}

func (pc ProviderCookie) Set(w http.ResponseWriter) error {
	if err := CookieNameEmailListID.SetEncrypted(w, pc.EmailListID); err != nil {
		return err
	}
	if err := CookieNameProviderName.SetEncrypted(w, string(pc.ProviderName)); err != nil {
		return err
	}
	if err := CookieNameOutputIDs.SetEncrypted(w, strings.Join(pc.OutputIDs, outputCookieDelim)); err != nil {
		return err
	}
	if err := CookieNameRedirectURL.SetEncrypted(w, pc.RedirectUrl); err != nil {
		return err
	}
	if err := CookieNameCreatedAt.SetEncrypted(w, pc.CreatedAt.Format(timestampFormat)); err != nil {
		return err
	}
	return nil
}

func (cn CookieName) encrypt() (string, error) {
	cookieSecret := os.Getenv(EnvCookieSecret)
	if cookieSecret == "" {
		return "", missingEnv(EnvCookieSecret)
	}
	return Encrypt(cookieSecret, string(cn))
}

func (cn CookieName) DecryptFrom(r *http.Request) (string, error) {
	cookieSecret := os.Getenv(EnvCookieSecret)
	if cookieSecret == "" {
		return "", missingEnv(EnvCookieSecret)
	}

	encryptedCookieName, err := cn.encrypt()
	if err != nil {
		return "", err
	}

	cookie, err := r.Cookie(encryptedCookieName)
	if err != nil {
		return "", err
	}

	return Decrypt(cookieSecret, cookie.Value)
}

func (cn CookieName) SetEncrypted(w http.ResponseWriter, value string) error {
	cookieSecret := os.Getenv(EnvCookieSecret)
	if cookieSecret == "" {
		return missingEnv(EnvCookieSecret)
	}

	encryptedCookieName, err := cn.encrypt()
	if err != nil {
		return err
	}

	encryptedValue, err := Encrypt(cookieSecret, value)
	if err != nil {
		return err
	}

	setCookie(w, encryptedCookieName, encryptedValue)
	return nil
}

func (pc ProviderCookie) Handle(pr ProviderResult) error {
	emailList, err := storage.GetEmailListByID(pc.EmailListID)
	if err != nil {
		return err
	}
	userID := emailList.UserID

	go func() {
		for _, outputID := range pc.OutputIDs {
			output, err := storage.GetOutputByIDAndUserID(outputID, userID)
			if err != nil {
				continue
			}
			output.Handle(pr.EmailAddr, pr.Name)
		}
	}()

	cr := SubscriberCreationReq{
		EmailListID:        pc.EmailListID,
		UserID:             userID,
		SourceProviderName: pc.ProviderName,
		Name:               pr.Name,
		EmailAddr:          pr.EmailAddr,
	}

	_, err = storage.InsertNewSubscriber(cr)
	return err
}

func setCookie(w http.ResponseWriter, name string, value string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   cookieMaxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

func clearCookie(w http.ResponseWriter, name string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Expires immediately
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}
