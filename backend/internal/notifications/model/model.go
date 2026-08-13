package model

type Status string

const (
	StatusSent    Status = "sent"
	StatusSkipped Status = "skipped"
)

type Message struct {
	Recipient string
	Subject   string
	TextBody  string
	HTMLBody  string
}

type DispatchResult struct {
	Status    Status
	Energy    int
	Threshold *int
}
