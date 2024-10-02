package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/EricFrancis12/stripol"
	"github.com/resend/resend-go/v2"
	sendinblue "github.com/sendinblue/APIv3-go-library/v2/lib"
)

func makeOutput(
	id string,
	userID string,
	outputName OutputName,
	listID string,
	param1 string,
	param2 string,
	param3 string,
	createdAt time.Time,
	updatedAt time.Time,
) Output {
	switch outputName {
	case OutputNameAWeber:
		return AWeberOutput{
			ID:         id,
			UserID:     userID,
			ListID:     listID,
			AdTracking: param1,
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
		}
	case OutputNameBrevo:
		return BrevoOutput{
			ID:         id,
			UserID:     userID,
			ListID:     listID,
			ExternalID: param1,
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
		}
	case OutputNameResend:
		return ResendOutput{
			ID:         id,
			UserID:     userID,
			AudienceID: listID,
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
		}
	case OutputNameTelegram:
		return TelegramOutput{
			ID:        id,
			UserID:    userID,
			ChatID:    listID,
			MsgFmt:    param1,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
	case OutputNameWebhook:
		return WebhookOutput{
			ID:        id,
			UserID:    userID,
			UrlFmt:    param1,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
	}
	return nil
}

func makeOutputsData(outputs []Output) OutputsData {
	od := make(OutputsData)

	for _, output := range outputs {
		if output == nil {
			continue
		}

		on := output.OutputName()
		_, ok := od[on]
		if ok {
			od[on] = append(od[on], output)
		} else {
			od[on] = []Output{output}
		}
	}

	return od
}

func (ao AWeberOutput) OutputName() OutputName {
	return OutputNameAWeber
}

func (ao AWeberOutput) GetUserID() string {
	return ao.UserID
}

func (ao AWeberOutput) Handle(emailAddr string, name string) error {
	formData := url.Values{}

	formData.Set(FormFieldListName, ao.ListID)
	formData.Set(FormFieldName, name)
	formData.Set(FormFieldEmail, emailAddr)
	if ao.AdTracking != "" {
		formData.Set(FormFieldAdTracking, ao.AdTracking)
	}

	encodedFormData := formData.Encode()

	// TODO: Find new way of adding lead to aweber list (other than posting to addlead.pl),
	// because this endpoint is possibly blocking traffic from cloud providers.
	req, err := http.NewRequest(
		http.MethodPost,
		"https://www.aweber.com/scripts/addlead.pl",
		bytes.NewBufferString(encodedFormData),
	)
	if err != nil {
		return err
	}
	req.Header.Add(HTTPHeaderAcceptEncoding, ContentTypeApplicationXwwwFormUrlEncoded)
	req.Header.Add(HTTPHeaderContentType, ContentTypeApplicationXwwwFormUrlEncoded)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("received %d status code", resp.StatusCode)
	}

	return nil
}

func (bo BrevoOutput) OutputName() OutputName {
	return OutputNameBrevo
}

func (bo BrevoOutput) GetUserID() string {
	return bo.UserID
}

func (bo BrevoOutput) Handle(emailAddr string, name string) error {
	brevoApiKey := os.Getenv(EnvBrevoApiKey)
	if brevoApiKey == "" {
		return missingEnv(EnvBrevoApiKey)
	}

	cfg := sendinblue.NewConfiguration()
	cfg.AddDefaultHeader("api-key", brevoApiKey)

	if bo.ListID == "" {
		return fmt.Errorf("listID cannot be empty")
	}
	i, err := strconv.Atoi(bo.ListID)
	if err != nil {
		return err
	}
	listID := int64(i)

	sib := sendinblue.NewAPIClient(cfg)

	contact := sendinblue.CreateContact{
		Email: emailAddr,
		Attributes: map[string]interface{}{
			"FIRSTNAME": name,
			"EXT_ID":    bo.ExternalID,
		},
		ListIds: []int64{
			listID,
		},
	}

	if _, _, err := sib.ContactsApi.CreateContact(context.Background(), contact); err != nil {
		return err
	}

	return nil
}

func (ro ResendOutput) OutputName() OutputName {
	return OutputNameResend
}

func (ro ResendOutput) GetUserID() string {
	return ro.UserID
}

func (ro ResendOutput) Handle(emailAddr string, name string) error {
	resendApiKey := os.Getenv(EnvResendApiKey)
	if resendApiKey == "" {
		return missingEnv(EnvResendApiKey)
	}

	client := resend.NewClient(resendApiKey)

	params := &resend.CreateContactRequest{
		Email:        emailAddr,
		FirstName:    name,
		LastName:     "",
		Unsubscribed: false,
		AudienceId:   ro.AudienceID,
	}
	_, err := client.Contacts.Create(params)
	if err != nil {
		return err
	}

	return nil
}

func (to TelegramOutput) OutputName() OutputName {
	return OutputNameTelegram
}

func (to TelegramOutput) GetUserID() string {
	return to.UserID
}

func (to TelegramOutput) StripolMap(emailAddr string, name string) map[string]string {
	return emailAddrAndNameStripolMap(emailAddr, name)
}

func (to TelegramOutput) Handle(emailAddr string, name string) error {
	telegramBotID := os.Getenv(EnvTelegramBotID)
	if telegramBotID == "" {
		return missingEnv(EnvTelegramBotID)
	}

	si := stripol.New(stripolLeftDelim, stripolRightDelim)
	si.RegisterVars(to.StripolMap(emailAddr, name))
	msg := si.Eval(to.MsgFmt)

	return SendMessageToTelegramChannel(telegramBotID, to.ChatID, msg)
}

func SendMessageToTelegramChannel(botID string, chatId string, message string) error {
	fdm := make(FormDataMap)
	fdm[FormFieldTelegramChatID] = strings.NewReader(chatId)
	fdm[FormFieldText] = strings.NewReader(message)
	return fdm.Upload(TelegramAPIMessageUrl(botID))
}

func (wo WebhookOutput) OutputName() OutputName {
	return OutputNameWebhook
}

func (wo WebhookOutput) GetUserID() string {
	return wo.UserID
}

func (wo WebhookOutput) StripolMap(emailAddr string, name string) map[string]string {
	return emailAddrAndNameStripolMap(emailAddr, name)
}

func (wo WebhookOutput) Handle(emailAddr string, name string) error {
	si := stripol.New(stripolLeftDelim, stripolRightDelim)
	si.RegisterVars(wo.StripolMap(url.QueryEscape(emailAddr), url.QueryEscape(name)))
	_url := si.Eval(wo.UrlFmt)

	_, err := http.Get(_url)
	return err
}

func emailAddrAndNameStripolMap(emailAddr string, name string) map[string]string {
	return map[string]string{
		StrIpolEmailAddr: emailAddr,
		StrIpolName:      name,
	}
}
