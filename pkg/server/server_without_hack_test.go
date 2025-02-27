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

type ServerWithoutHackTestSuite struct {
	suite.Suite
	server *Server
}

func (suite *ServerWithoutHackTestSuite) SetupTest() {
	suite.server = NewServer(false)
}

func TestServerWithoutHackTestSuite(t *testing.T) {
	suite.Run(t, new(ServerWithoutHackTestSuite))
}

func (suite *ServerWithoutHackTestSuite) serveHTTP(w http.ResponseWriter, r *http.Request) {
	suite.server.intHttpServer.Handler.ServeHTTP(w, r)
}

func (suite *ServerWithoutHackTestSuite) decodeJSON(w *httptest.ResponseRecorder, target any) {
	suite.Assert().Equal("application/json", w.Header().Get("Content-Type"))
	err := json.NewDecoder(w.Body).Decode(target)
	suite.Require().NoErrorf(err, "failed to decode JSON: %v", err)
}

func (suite *ServerWithoutHackTestSuite) newJSONRequest(method string, url string, body any) *http.Request {
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

func (suite *ServerWithoutHackTestSuite) TestDiscriminatorMapping() {
	body := `{
		"petClass": "MAMMAL",
		"species": "dog"
	}`
	req := suite.newJSONRequest(http.MethodPost, "/pets", body)
	w := httptest.NewRecorder()
	suite.serveHTTP(w, req)
	suite.Assert().Equal(http.StatusNotAcceptable, w.Code)
}
