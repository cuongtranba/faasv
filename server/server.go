package server

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/cuongtranba/faasv"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/rs/zerolog/log"
)

type GatewayServer struct {
	queue      faasv.Queue
	httpServer *http.Server
	ctx        context.Context
	middleware []func(http.Handler) http.Handler
}

func (server *GatewayServer) AddMiddleware(middleware ...func(http.Handler) http.Handler) {
	server.middleware = append(server.middleware, middleware...)
}

func (server *GatewayServer) middlewareApplier(handler http.Handler) http.Handler {
	for _, middleware := range server.middleware {
		handler = middleware(handler)
	}
	return handler
}

func (server *GatewayServer) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		log.Err(err).Msg("Failed to upgrade websocket")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	go func() {
		defer conn.Close()

		for {
			msg, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				log.Err(err).Msg("Failed to read client data")
				continue
			}
			req, err := ConvertWSMsgToNats(msg)
			if err != nil {
				log.Err(err).Msg("Failed to convert ws message to nats request")
				continue
			}
			log.Info().Msgf("Received request %s", req.Subject)
			res, err := server.queue.Request(r.Context(), req.Subject, req.Payload)
			if err != nil {
				log.Err(err).Msgf("Failed to process request %s", r.URL.Path)
				continue
			}
			msg, err = ConvertNatsResponseToWSMsg(res)
			if err != nil {
				log.Err(err).Msg("Failed to convert nats response to ws message")
				continue
			}

			err = wsutil.WriteServerMessage(conn, op, msg)
			if err != nil {
				log.Err(err).Msg("Failed to write server message")
				continue
			}
		}
	}()
}

func (g *GatewayServer) handleHttp(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	req, err := ConvertHTTPRequestToNats(r.Body)
	if err != nil {
		log.Err(err).Msg("Failed to convert http request to nats request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Info().Msgf("Received request %s", req.Subject)
	w.Header().Set("Content-Type", "application/json")

	res, err := g.queue.Request(r.Context(), req.Subject, req.Payload)
	if err != nil {
		log.Err(err).Msgf("Failed to process request %s", r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Error: "something_went_wrong"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{Payload: res})
}

func (g *GatewayServer) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/ws") {
			g.handleWebsocket(w, r)
		} else {
			g.handleHttp(w, r)
		}
	})
}

func NewGatewayServer(ctx context.Context, queue faasv.Queue, addr string) *GatewayServer {
	httpServer := &http.Server{
		Addr: addr,
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
	server.httpServer.Handler = server.middlewareApplier(server.handler())
	return server.httpServer.ListenAndServe()
}

func (server *GatewayServer) Url() string {
	return server.httpServer.Addr
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
