package server_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"             // Better HTTP API.
	"github.com/stretchr/testify/assert" // More readable test assertions.

	"github.com/marccarre/kubernetes-deployment-strategies-workload/pkg/db/dbtest"
	"github.com/marccarre/kubernetes-deployment-strategies-workload/pkg/server"
)

const (
	lukeSkywalker = "{\"id\":1,\"firstName\":\"Luke\",\"familyName\":\"Skywalker\"}"
	obiWanKenobi  = "{\"id\":2,\"firstName\":\"Obi-Wan\",\"familyName\":\"Kenobi\"}"
)

func TestCreateAndReadUsers(t *testing.T) {
	database := dbtest.Setup(t)
	assert.NotNil(t, database)
	defer dbtest.Cleanup(t, database)
	server := server.New(database)

	req := get(t, "/users/1")
	resp := serve(req, server)
	assert.Equal(t, http.StatusNotFound, resp.Code)
	assert.Equal(t, "", body(t, resp.Body))

	req = post(t, "/users", "{\"firstName\":\"Luke\",\"familyName\":\"Skywalker\"}")
	resp = serve(req, server)
	assert.Equal(t, http.StatusCreated, resp.Code)
	assert.Equal(t, "/users/1", resp.Header().Get("Location"))
	assert.Equal(t, "", body(t, resp.Body))

	req = get(t, "/users/1")
	resp = serve(req, server)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, lukeSkywalker, body(t, resp.Body))

	req = post(t, "/users", "{\"firstName\":\"Obi-Wan\",\"familyName\":\"Kenobi\"}")
	resp = serve(req, server)
	assert.Equal(t, http.StatusCreated, resp.Code)
	assert.Equal(t, "/users/2", resp.Header().Get("Location"))
	assert.Equal(t, "", body(t, resp.Body))

	req = get(t, "/users/2")
	resp = serve(req, server)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, obiWanKenobi, body(t, resp.Body))

	req = get(t, "/users")
	resp = serve(req, server)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "["+lukeSkywalker+","+obiWanKenobi+"]", body(t, resp.Body))
}

func post(t *testing.T, uri, body string) *http.Request {
	return newRequest(t, "POST", uri, bytes.NewReader([]byte(body)))
}

func get(t *testing.T, uri string) *http.Request {
	return newRequest(t, "GET", uri, nil)
}

func newRequest(t *testing.T, verb, uri string, body io.Reader) *http.Request {
	req, err := http.NewRequest(verb, uri, body)
	assert.NoError(t, err)
	return req
}

func serve(req *http.Request, server *server.HTTPServer) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	router := mux.NewRouter()
	server.RegisterRoutes(router)
	router.ServeHTTP(resp, req)
	return resp
}

func body(t *testing.T, body *bytes.Buffer) string {
	bytes, err := ioutil.ReadAll(body)
	assert.NoError(t, err)
	return string(bytes)
}
