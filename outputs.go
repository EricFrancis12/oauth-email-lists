package main

import (
	"fmt"

	"github.com/resend/resend-go/v2"
)

// TODO: remove apiKey from the output models (and db), and instead have it be an env variable
func makeOutput(id string, userID string, outputName OutputName, apiKey string, listID string) Output {
	switch outputName {
	case OutputNameAWeber:
		return AWeberOutput{
			ID:     id,
			UserID: userID,
			ApiKey: apiKey,
			ListID: listID,
		}
	case OutputNameResend:
		return ResendOutput{
			ID:         id,
			UserID:     userID,
			ApiKey:     apiKey,
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
	client := resend.NewClient(ro.ApiKey)

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
