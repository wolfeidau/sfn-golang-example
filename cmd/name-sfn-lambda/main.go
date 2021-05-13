package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/aws/aws-lambda-go/lambda"
	lmw "github.com/wolfeidau/lambda-go-extras/middleware"
	"github.com/wolfeidau/lambda-go-extras/middleware/raw"
	zlog "github.com/wolfeidau/lambda-go-extras/middleware/zerolog"
)

var (
	version = "unknown"

	cli struct {
		Version         kong.VersionFlag
		RawEventLogging bool   `help:"Enable raw event logging." env:"RAW_EVENT_LOGGING"`
		Debug           bool   `help:"Enable debug logging." env:"DEBUG"`
		Stage           string `help:"The development stage." env:"STAGE"`
		Branch          string `help:"The git branch this code originated." env:"BRANCH"`
	}
)

// NewNameValidationError returns a new name validation error which is used to match errors in the step function.
func NewNameValidationError(msg string) error {
	return &NameValidationError{msg}
}

// NameValidationError Used to provide a custom error to step functions, one thing to note
// is the name of the error is what matters in the step function
type NameValidationError struct {
	msg string
}

// Error This function is required to implement an Error and the value is visible in the step which handles the error
func (e *NameValidationError) Error() string { return e.msg }

// SFNEvent generic top level wrapper for the step function event with the input free to be changed for each lambda
type SFNEvent struct {
	Path  string
	Input json.RawMessage
}

// NameParams used to pull out the params used by the SFN task lambdas
type NameParams struct {
	Name string `json:"name,omitempty"`
}

func parseNameParams(payload []byte) (*NameParams, error) {
	params := new(NameParams)

	err := json.Unmarshal(payload, params)
	if err != nil {
		return nil, err
	}

	return params, nil
}

func main() {
	kong.Parse(&cli,
		kong.Vars{"version": version}, // bind a var for version
	)

	// build up a list of fields which will be included in all log messages
	flds := lmw.FieldMap{"version": version}

	ch := lmw.New(
		zlog.New(zlog.Fields(flds)), // assign a logger and bind it in the context
	)

	if cli.RawEventLogging {
		ch.Use(raw.New(raw.Fields(flds))) // if raw event logging is enabled dump everything to the log in and out
	}

	// register our lambda handler with the middleware configured
	lambda.StartHandler(ch.ThenFunc(processEvent))
}

func processEvent(ctx context.Context, payload []byte) ([]byte, error) {

	evt := new(SFNEvent)

	err := json.Unmarshal(payload, evt)
	if err != nil {
		return nil, err
	}

	// dispatch based on path
	switch evt.Path {
	case "ValidateName":
		return handleValidateName(evt)
	case "InvalidName":
		return handleInvalidName(evt)
	case "GreetName":
		return handleGreetName(evt)
	}

	return nil, errors.New("woops")
}

func handleInvalidName(evt *SFNEvent) ([]byte, error) {
	params, err := parseNameParams(evt.Input)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf(`{"error": "validate name failed for: %s"}`, params.Name)), nil
}

func handleValidateName(evt *SFNEvent) ([]byte, error) {
	params, err := parseNameParams(evt.Input)
	if err != nil {
		return nil, err
	}

	if params.Name == "Mark" {
		return evt.Input, nil
	}

	return nil, NewNameValidationError("not the name I am looking for")
}

func handleGreetName(evt *SFNEvent) ([]byte, error) {
	params := new(NameParams)

	err := json.Unmarshal(evt.Input, params)
	if err != nil {
		return nil, err
	}

	return []byte(fmt.Sprintf(`{"message": "hello %s"}`, params.Name)), nil
}
