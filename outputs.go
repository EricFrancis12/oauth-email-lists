package main

import "fmt"

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
			ID:     id,
			UserID: userID,
			ApiKey: apiKey,
			ListID: listID,
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
	// TODO: ...
	fmt.Println("~ ResendOutput.Handle() called")
	return nil
}
