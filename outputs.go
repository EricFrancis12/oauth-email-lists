package main

import (
	"fmt"
	"os"
	"time"

	"github.com/resend/resend-go/v2"
)

func makeOutput(id string, userID string, outputName OutputName, listID string, createdAt time.Time) Output {
	switch outputName {
	case OutputNameAWeber:
		return AWeberOutput{
			ID:        id,
			UserID:    userID,
			ListID:    listID,
			CreatedAt: createdAt,
		}
	case OutputNameResend:
		return ResendOutput{
			ID:         id,
			UserID:     userID,
			AudienceID: listID,
			CreatedAt:  createdAt,
		}
	}
	return nil
}

func (ao AWeberOutput) OutputName() OutputName {
	return OutputNameAWeber
}

func (ao AWeberOutput) Handle(emailAddr string, name string) error {
	// TODO: ...
	fmt.Println("~ AWeberOutput.Handle() called")
	return nil
}

func (ro ResendOutput) OutputName() OutputName {
	return OutputNameResend
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
