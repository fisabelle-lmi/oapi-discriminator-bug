package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fisabelle-lmi/oapi-discriminator-bug/pkg/api/common"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
	"log/slog"
	"net/http"
	"os"

	privateApi "github.com/fisabelle-lmi/oapi-discriminator-bug/pkg/api/private"
	"github.com/gorilla/mux"
)

type Server struct {
	intHttpServer *http.Server
	extHttpServer *http.Server
}

type Middleware func(next http.Handler) http.HandlerFunc

func toPrivateApiMiddleware(mw Middleware) privateApi.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return mw(next)
	}
}

func NewServer(withHack bool) *Server {
	slog.Info("creating new server")
	// Regular router
	router := mux.NewRouter().StrictSlash(true)
	router.NotFoundHandler = notFoundHandler

	handler := func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			next.ServeHTTP(w, req)
		}
	}

	privateApiServer := NewPrivateApiServer()
	strictInternalServerOptions := privateApi.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  inputErrorHandler,
		ResponseErrorHandlerFunc: internalErrorHandler,
	}

	internalSwagger, err := privateApi.GetSwagger()
	if err != nil {
		panic(fmt.Errorf("unable to load swagger spec: %w", err))
	}
	// remove servers from the spec so that the validator uses the request URL
	internalSwagger.Servers = nil

	// Hack for discriminator bug
	// The swagger generator does not correctly prefix discriminator mapping with the schema name.
	// The swagger generator incorrect keeps the enum mapping to the "#/components/schemas/Mammal" instead of  "#/components/schemas/common_Mammal
	// which is the one the validator expects.
	if withHack {
		internalSwagger.Components.Schemas["common_Pet"].Value.Discriminator.Mapping["MAMMAL"] = "#/components/schemas/common_Mammal"
		internalSwagger.Components.Schemas["common_Pet"].Value.Discriminator.Mapping["AMPHIBIAN"] = "#/components/schemas/common_Amphibian"
	}

	privateApiStrictHandler := privateApi.NewStrictHandlerWithOptions(privateApiServer, nil, strictInternalServerOptions)

	intHandlerOptions := privateApi.GorillaServerOptions{
		BaseRouter:       router,
		ErrorHandlerFunc: inputErrorHandler,
		Middlewares: []privateApi.MiddlewareFunc{
			nethttpmiddleware.OapiRequestValidatorWithOptions(internalSwagger, &nethttpmiddleware.Options{
				ErrorHandlerWithOpts: func(w http.ResponseWriter, message string, statusCode int, opts nethttpmiddleware.ErrorHandlerOpts) {
					inputErrorHandler(w, opts.Request, errors.New(message))
				},
			}),
			toPrivateApiMiddleware(handler),
		},
	}

	privateApi.HandlerWithOptions(privateApiStrictHandler, intHandlerOptions)

	return &Server{
		intHttpServer: &http.Server{
			Addr:    ":8080",
			Handler: router,
		},
	}
}

func (s *Server) Start() {
	go func() {
		slog.Info("starting internal http server on port 8080...")
		if err := s.intHttpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server error", "error", err)
			os.Exit(1)
		}
	}()
}

var notFoundHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

var inputErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("input error", "error", err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(common.N400{ErrorCode: "CONSTRAINT_VIOLATION", Message: err.Error()})
}

var internalErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("internal error", "error", err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(common.N500{ErrorCode: "INTERNAL_SERVER_ERROR", Message: err.Error()})
}
