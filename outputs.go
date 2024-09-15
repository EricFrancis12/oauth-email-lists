package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/resend/resend-go/v2"
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
			ID:        id,
			UserID:    userID,
			ListID:    listID,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
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
	// TODO: ...
	fmt.Println("~ AWeberOutput.Handle() called")
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

func (to TelegramOutput) Handle(emailAddr string, name string) error {
	telegramBotID := os.Getenv(EnvTelegramBotID)
	if telegramBotID == "" {
		return missingEnv(EnvTelegramBotID)
	}

	si := NewStrIpol(strIpolLeftDelim, strIpolRightDelim)
	si.RegisterVars(to.StrIpolMap(emailAddr, name))
	msg := si.Eval(to.MsgFmt)

	return SendMessageToTelegramChannel(telegramBotID, to.ChatID, msg)
}

func SendMessageToTelegramChannel(botID string, chatId string, message string) error {
	fdm := make(FormDataMap)
	fdm[FormFieldTelegramChatID] = strings.NewReader(chatId)
	fdm[FormFieldText] = strings.NewReader(chatId)
	return fdm.Upload(TelegramAPIMessageUrl(botID))
}
