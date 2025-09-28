package api

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
)

type ErrorCode int

const (
	NoError ErrorCode = iota
	_
	ErrInvalidService
	ErrInvalidMethod
	ErrAuthenticationFailed
	ErrInvalidFormat
	ErrInvalidParameters
	ErrInvalidResource
	ErrOperationFailed
	ErrInvalidSessionKey
	ErrInvalidAPIKey
	ErrServiceOffline
	ErrSubscribersOnly
	ErrInvalidMethodSignature
	ErrUnauthorizedToken
	ErrItemNotStreamable
	ErrServiceUnavailable
	ErrUserNotLoggedIn
	ErrTrialExpired
	_
	ErrNotEnoughContent
	ErrNotEnoughMembers
	ErrNotEnoughFans
	ErrNotEnoughNeighbours
	ErrNoPeakRadio
	ErrRadioNotFound
	ErrAPIKeySuspended
	ErrDeprecated
	_
	ErrRateLimitExceeded
)

const (
	ErrAPIKeyMissing ErrorCode = iota + 100
	ErrSecretRequired
	ErrSessionRequired
)

const (
	APIKeyMissingMessage   = "API Key is missing"
	SecretRequiredMessage  = "Method requires API secret"
	SessionRequiredMessage = "Method requires user authentication (session key)"
)

type LFMWrapper struct {
	XMLName  xml.Name `xml:"lfm"`
	Status   string   `xml:"status,attr"`
	InnerXML []byte   `xml:",innerxml"`
}

func (lf *LFMWrapper) Empty() bool                      { return len(lf.InnerXML) == 0 }
func (lf *LFMWrapper) StatusOK() bool                   { return lf.Status == "ok" }
func (lf *LFMWrapper) StatusFailed() bool               { return lf.Status == "failed" }
func (lf *LFMWrapper) UnmarshalInnerXML(dest any) error { return xml.Unmarshal(lf.InnerXML, dest) }
func (lf *LFMWrapper) UnwrapError() (*LastFMError, error) {
	if lf.StatusOK() {
		return nil, nil
	}
	var lferr LastFMError
	if err := lf.UnmarshalInnerXML(&lferr); err != nil {
		return nil, err
	}
	if lferr.HasErrorCode() {
		return &lferr, nil
	}
	return nil, errors.New("no error code in response")
}

type LastFMError struct {
	Code      ErrorCode `xml:"code,attr"`
	Message   string    `xml:",chardata"`
	httpError *HTTPError
}

func NewLastFMError(code ErrorCode, message string) *LastFMError {
	return &LastFMError{Code: code, Message: message}
}
func (e *LastFMError) Error() string { return fmt.Sprintf("Last.fm Error: %d - %s", e.Code, e.Message) }
func (e *LastFMError) Is(target error) bool {
	if t, ok := target.(*LastFMError); ok {
		return e.IsCode(t.Code)
	}
	return false
}
func (e *LastFMError) Unwrap() error { return e.httpError }
func (e *LastFMError) WrapHTTPError(httpError *HTTPError) *LastFMError {
	e.httpError = httpError
	return e
}
func (e *LastFMError) WrapResponse(res *http.Response) *LastFMError {
	return e.WrapHTTPError(NewHTTPError(res))
}
func (e LastFMError) IsCode(code ErrorCode) bool { return e.Code == code }
func (e LastFMError) HasErrorCode() bool         { return e.Code != NoError }
func (e LastFMError) ShouldRetry() bool {
	return e.Code == ErrOperationFailed || e.Code == ErrServiceUnavailable || e.Code == ErrRateLimitExceeded
}

type HTTPError struct {
	StatusCode int
	Message    string
}

func NewHTTPError(res *http.Response) *HTTPError {
	if res == nil {
		return &HTTPError{StatusCode: http.StatusInternalServerError, Message: "nil response"}
	}
	return &HTTPError{StatusCode: res.StatusCode, Message: http.StatusText(res.StatusCode)}
}
func (e *HTTPError) Error() string { return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message) }
func (e *HTTPError) Is(target error) bool {
	if t, ok := target.(*HTTPError); ok {
		return e.StatusCode == t.StatusCode
	}
	return false
}
