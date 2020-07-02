// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package utils

import (
	"encoding/json"
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
)

type httpError struct {
	cause  error
	status int
}

func (e *httpError) Error() string {
	return e.cause.Error()
}

// HTTPError create an error with http status code.
func HTTPError(cause error, status int) error {
	return &httpError{
		cause:  cause,
		status: status,
	}
}

// BadRequest convenience method to create http bad request error.
func BadRequest(cause error) error {
	return &httpError{
		cause:  cause,
		status: http.StatusBadRequest,
	}
}

// Forbidden convenience method to create http forbidden error.
func Forbidden(cause error) error {
	return &httpError{
		cause:  cause,
		status: http.StatusForbidden,
	}
}

// HandlerFunc like http.HandlerFunc, bu it returns an error.
// If the returned error is httpError type, httpError.status will be responded,
// otherwise http.StatusInternalServerError responded.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// WrapHandlerFunc convert HandlerFunc to http.HandlerFunc.
func WrapHandlerFunc(f HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			if he, ok := err.(*httpError); ok {
				if he.cause != nil {
					http.Error(w, he.cause.Error(), he.status)
				} else {
					w.WriteHeader(he.status)
				}
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}

// content types
const (
	JSONContentType        = "application/json; charset=utf-8"
	OctetStreamContentType = "application/octet-stream"
)

// ParseJSON parse a JSON object using strict mode.
func ParseJSON(r io.Reader, v interface{}) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

// WriteJSON reponse a object in JSON enconding.
func WriteJSON(w http.ResponseWriter, obj interface{}) error {
	//data, err := json.Marshal(obj)
	data, err := json.MarshalIndent(obj,"","\t")
	if err != nil {
		return HTTPError(err, 500)
	}
	w.Header().Set("Content-Type", JSONContentType)
	w.Write(data)
	return nil
}

func CreateWs(w http.ResponseWriter, req *http.Request) *websocket.Conn {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	ws, _  := upgrader.Upgrade(w, req, nil)
	return  ws
}

func WriteJSONWs(w http.ResponseWriter, obj interface{}, ws *websocket.Conn) error {
	defer ws.Close()
	//data, err := json.Marshal(obj)
	data, err := json.MarshalIndent(obj,"","\t")
	if err != nil {
		return HTTPError(err, 500)
	}
	w.Header().Set("Content-Type", JSONContentType)
	if ws != nil {
		ws.WriteMessage(websocket.TextMessage, data)
		ws.WriteMessage(websocket.CloseMessage, []byte(""))
	}
	return nil
}

func WriteTo(w http.ResponseWriter, req *http.Request, obj interface{}) error {
	if polo.ApiProtocol==0 || polo.ApiProtocol==1{
		return WriteJSON(w, obj)
	}else {
		ws := CreateWs(w, req)
		return WriteJSONWs(w, obj, ws)
	}
}

// M shortcut for type map[string]interface{}.
type M map[string]interface{}
