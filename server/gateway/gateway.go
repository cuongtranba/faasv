package http

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/cuongtranba/faasv"
	"github.com/rs/zerolog/log"
)

type GatewayServer struct {
	queue      faasv.Queue
	httpServer *http.Server
	ctx        context.Context
}

type ServerResponse struct {
	Payload any    `json:"payload"`
	Error   string `json:"error"`
}

// convert http request to nats request
func natHandler(queue faasv.Queue) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		subject := strings.TrimPrefix(r.URL.Path, "/") // convert request path to nat subject

		log.Info().Msgf("Received request %s", subject)
		w.Header().Set("Content-Type", "application/json")
		var natRequest any
		err := json.NewDecoder(r.Body).Decode(&natRequest)
		if err != nil {
			log.Error().Msgf("Decode request error: %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ServerResponse{Error: "something_went_wrong"})
		}
		res, err := queue.Request(r.Context(), subject, natRequest)
		if err != nil {
			log.Err(err).Msgf("Failed to process request %s", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ServerResponse{Error: "something_went_wrong"})
			return
		}
		json.NewEncoder(w).Encode(ServerResponse{Payload: res})
	})
}

func NewGatewayServer(ctx context.Context, queue faasv.Queue, addr string) *GatewayServer {
	httpServer := &http.Server{
		Addr:    addr,
		Handler: natHandler(queue),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
	return &GatewayServer{
		queue:      queue,
		httpServer: httpServer,
		ctx:        ctx,
	}
}

func (server *GatewayServer) Start() error {
	log.Info().Msgf("Starting gateway server %s", server.httpServer.Addr)
	return server.httpServer.ListenAndServe()
}

func (server *GatewayServer) Stop(d time.Duration) error {
	log.Info().Msgf("Stopping gateway server %s", server.httpServer.Addr)
	ctx, cancel := context.WithTimeout(server.ctx, d)
	defer cancel()
	if err := server.httpServer.Shutdown(ctx); err != nil {
		return err
	}
	log.Info().Msgf("Stopped gateway server %s", server.httpServer.Addr)
	return nil
}
