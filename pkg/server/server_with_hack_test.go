package server

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type ServerTestSuite struct {
	suite.Suite
	server *Server
}

func (suite *ServerTestSuite) SetupTest() {
	suite.server = NewServer(true)
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

func (suite *ServerTestSuite) serveHTTP(w http.ResponseWriter, r *http.Request) {
	suite.server.intHttpServer.Handler.ServeHTTP(w, r)
}

func (suite *ServerTestSuite) decodeJSON(w *httptest.ResponseRecorder, target any) {
	suite.Assert().Equal("application/json", w.Header().Get("Content-Type"))
	err := json.NewDecoder(w.Body).Decode(target)
	suite.Require().NoErrorf(err, "failed to decode JSON: %v", err)
}

func (suite *ServerTestSuite) newJSONRequest(method string, url string, body any) *http.Request {
	var req *http.Request
	if body != nil {
		if s, ok := body.(string); ok {
			req = httptest.NewRequest(method, url, strings.NewReader(s))
		} else {
			b, err := json.Marshal(body)
			suite.Require().NoErrorf(err, "failed to marshal JSON: %v", err)
			req = httptest.NewRequest(method, url, bytes.NewReader(b))
		}
	} else {
		req = httptest.NewRequest(method, url, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

func (suite *ServerTestSuite) TestDiscriminatorMapping() {
	body := `{
		"petClass": "MAMMAL",
		"species": "dog"
	}`
	req := suite.newJSONRequest(http.MethodPost, "/pets", body)
	w := httptest.NewRecorder()
	suite.serveHTTP(w, req)
	suite.Assert().Equal(http.StatusNotAcceptable, w.Code)
}
