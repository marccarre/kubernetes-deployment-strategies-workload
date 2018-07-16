package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"         // Better HTTP API.
	log "github.com/sirupsen/logrus" // Better Logging.

	"github.com/marccarre/kubernetes-deployment-strategies-workload/pkg/db"
	"github.com/marccarre/kubernetes-deployment-strategies-workload/pkg/domain"
)

// HTTPServer is an HTTP server reading users from the configured database.
type HTTPServer struct {
	db db.DB
}

// New creates a new HTTP server.
func New(db db.DB) *HTTPServer {
	return &HTTPServer{
		db: db,
	}
}

// RegisterRoutes registers the users API HTTP routes to the provided mux.Router.
func (server *HTTPServer) RegisterRoutes(router *mux.Router) {
	for _, route := range server.routes() {
		router.Handle(route.Path, route.Handler).Methods(route.Method).Name(route.Name)
	}
}

type route struct {
	Name    string           `json:"-"`
	Method  string           `json:"method"`
	Path    string           `json:"path"`
	Handler http.HandlerFunc `json:"-"`
}

func (server HTTPServer) routes() []route {
	return []route{
		{"routes", "GET", "/", server.Routes},
		{"healthz", "GET", "/healthz", server.CheckHealth},
		{"users", "POST", "/users", server.CreateUserHandler},
		{"users", "GET", "/users", server.ReadUsersHandler},
		{"users_id", "GET", "/users/{id:[0-9]+}", server.ReadUserByIDHandler},
	}
}

// Routes lists this server's endpoints.
func (server HTTPServer) Routes(resp http.ResponseWriter, req *http.Request) {
	logger := log.WithField("method", req.Method).WithField("path", req.URL.Path)
	bytes, err := json.Marshal(server.routes())
	if err != nil {
		writeError(resp, logger, err, "failed to serialise routes as JSON", http.StatusInternalServerError)
		return
	}
	writeResponse(resp, logger, bytes)
}

// CheckHealth checks the health of this server.
func (server HTTPServer) CheckHealth(resp http.ResponseWriter, req *http.Request) {
	logger := log.WithField("method", req.Method).WithField("path", req.URL.Path)
	err := server.db.Ping(req.Context())
	if err != nil {
		writeError(resp, logger, err, "health check failed", http.StatusInternalServerError)
		return
	}
	resp.WriteHeader(http.StatusNoContent)
}

// CreateUserHandler stores the provided user.
func (server HTTPServer) CreateUserHandler(resp http.ResponseWriter, req *http.Request) {
	logger := log.WithField("method", req.Method).WithField("path", req.URL.Path)
	json, err := ioutil.ReadAll(req.Body)
	if err != nil {
		writeError(resp, logger, err, "failed to read request's body", http.StatusInternalServerError)
		return
	}
	user, err := domain.UnmarshalUser(json)
	if err != nil {
		writeError(resp, logger, err, "failed to deserialise user", http.StatusInternalServerError)
		return
	}
	id, err := server.db.CreateUser(req.Context(), user)
	if err != nil {
		writeError(resp, logger, err, "failed to create user", http.StatusInternalServerError)
		return
	}
	resp.Header().Set("Location", fmt.Sprintf("/users/%v", id))
	resp.WriteHeader(http.StatusCreated)
}

// ReadUsersHandler returns all stored users.
func (server HTTPServer) ReadUsersHandler(resp http.ResponseWriter, req *http.Request) {
	logger := log.WithField("method", req.Method).WithField("path", req.URL.Path)
	users, err := server.db.ReadUsers(req.Context())
	if err != nil {
		writeError(resp, logger, err, "failed to read users", http.StatusInternalServerError)
		return
	}
	bytes, err := json.Marshal(users)
	if err != nil {
		writeError(resp, logger, err, "failed to serialise users as JSON", http.StatusInternalServerError)
		return
	}
	writeResponse(resp, logger, bytes)
}

// ReadUserByIDHandler return the stored user corresponding to the provided ID.
func (server HTTPServer) ReadUserByIDHandler(resp http.ResponseWriter, req *http.Request) {
	idStr := mux.Vars(req)["id"]
	logger := log.WithField("method", req.Method).WithField("path", req.URL.Path).WithField("id", idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(resp, logger, err, "invalid ID", http.StatusBadRequest)
		return
	}
	user, err := server.db.ReadUserByID(req.Context(), id)
	if err != nil {
		if err == db.ErrNotFound {
			writeError(resp, logger, err, "failed to read user", http.StatusNotFound)
			return
		}
		writeError(resp, logger, err, "failed to read user", http.StatusInternalServerError)
		return
	}
	bytes, err := user.Marshal()
	if err != nil {
		writeError(resp, logger, err, "failed to serialise user as JSON", http.StatusInternalServerError)
		return
	}
	writeResponse(resp, logger, bytes)
}

func writeError(resp http.ResponseWriter, logger *log.Entry, err error, message string, status int) {
	logger.WithField("err", err).Error(message)
	resp.WriteHeader(status)
}

func writeResponse(resp http.ResponseWriter, logger *log.Entry, bytes []byte) {
	resp.WriteHeader(http.StatusOK)
	resp.Header().Set("Content-Type", "application/json")
	bytesWritten, err := resp.Write(bytes)
	logger = logger.WithField("bytesWritten", bytesWritten).WithField("bytes", len(bytes))
	if len(bytes) != bytesWritten {
		logger.Warn("response was not fully written")
	}
	if err != nil {
		logger.WithField("err", err).Error("failed to write response")
	}
}
