package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/tracing"
	"github.com/mathieupost/jetflow/transport/jetstream"
	"github.com/nats-io/nats.go"
	natsjetstream "github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel"

	"github.com/mathieupost/jetflow/example/types"
	"github.com/mathieupost/jetflow/example/types/gen"
)

func main() {
	tp, shutdown, err := tracing.NewProvider("jaeger:4318")
	if err != nil {
		log.Fatal("new tracing provider", err.Error())
	}
	defer shutdown()
	otel.SetTracerProvider(tp)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nc, err := nats.Connect("nats")
	if err != nil {
		log.Fatal("connecting to NATS", err.Error())
	}

	js, err := natsjetstream.New(nc)
	if err != nil {
		log.Fatal("initializing JetStream instance", err.Error())
	}

	factoryMapping := gen.ProxyFactoryMapping()
	publisher := jetstream.NewPublisher(ctx, js)
	client := jetflow.NewClient(factoryMapping, publisher)

	r := chi.NewRouter()
	r.Route("/users/{user}", func(r chi.Router) {
		r.Use(Operator[types.User](client, "user"))

		r.Route("/TransferBalance/{user2}/{amount}", func(r chi.Router) {
			r.Use(Operator[types.User](client, "user2"))
			r.Use(Integer("amount"))

			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				user := r.Context().Value("user").(types.User)
				user2 := r.Context().Value("user2").(types.User)
				amount := r.Context().Value("amount").(int)
				user.TransferBalance(r.Context(), user2, amount)
			})
		})
	})
	log.Println("Running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func Operator[O jetflow.Operator](client jetflow.OperatorClient, param string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, param)
			var operator O
			err := client.Find(r.Context(), id, &operator)
			if err != nil {
				http.Error(w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
				return
			}
			ctx := context.WithValue(r.Context(), param, operator)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Integer(param string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			str := chi.URLParam(r, param)
			res, err := strconv.Atoi(str)
			if err != nil {
				http.Error(w,
					fmt.Sprintf("'%s' is not a valid integer", param),
					http.StatusBadRequest)
				return
			}
			ctx := context.WithValue(r.Context(), param, res)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}