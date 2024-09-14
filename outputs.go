package main

import (
	"fmt"
	"os"

	"github.com/resend/resend-go/v2"
)

func makeOutput(id string, userID string, outputName OutputName, listID string) Output {
	switch outputName {
	case OutputNameAWeber:
		return AWeberOutput{
			ID:     id,
			UserID: userID,
			ListID: listID,
		}
	case OutputNameResend:
		return ResendOutput{
			ID:         id,
			UserID:     userID,
			AudienceID: listID,
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
