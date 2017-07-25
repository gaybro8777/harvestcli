package event

import (
	"encoding/json"
)

type LogType int

const (
	SEARCH LogType = iota
	CLICK
)

type LogLine struct {
	JSONPayload Payload `json:"jsonPayload"`
}

type Payload struct {
	Timestamp       json.Number `json:"timestamp"`
	Index           string      `json:"index"`
	AppID           string      `json:"appID"`
	QueryID         string      `json:"queryID"`
	UserID          string      `json:"userID"`
	Context         string      `json:"context"`
	Query           string      `json:"query"`
	QueryParameters string      `json:"queryParameters"`
	Position        json.Number `json:"position"`
	ObjectID        string      `json:"objectID"`
}

func GetLogType(logLine LogLine) LogType {
	// Index is a mandatory field for search event
	if len(logLine.JSONPayload.Index) > 0 {
		return SEARCH
	}
	return CLICK
}

func NewLogLine(payload []byte) (LogLine, error) {
	var logLine LogLine
	if err := json.Unmarshal(payload, &logLine); err != nil {
		return logLine, err
	}
	return logLine, nil
}

type SearchEvent struct {
	Timestamp       json.Number `json:"timestamp"`
	Index           string      `json:"index"`
	AppID           string      `json:"appID"`
	QueryID         string      `json:"queryID"`
	UserID          string      `json:"userID"`
	Context         string      `json:"context"`
	Query           string      `json:"query"`
	QueryParameters string      `json:"queryParameters"`
}
