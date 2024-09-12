package main

func makeOutput(id string, outputName OutputName, apiKey string, listID string) Output {
	switch outputName {
	case OutputNameAWeber:
		return AWeberOutput{
			ID:     id,
			ApiKey: apiKey,
			ListID: listID,
		}
	case OutputNameResend:
		return ResendOutput{
			ID:     id,
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
	return nil
}

func (ro ResendOutput) OutputName() OutputName {
	return OutputNameResend
}

func (ro ResendOutput) Handle(emailAddr string, name string) error {
	// TODO: ...
	return nil
}
