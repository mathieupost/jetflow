package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/tracing"
	"github.com/mathieupost/jetflow/transport/jetstream"
	"github.com/nats-io/nats.go"
	natsjetstream "github.com/nats-io/nats.go/jetstream"
	"github.com/pingcap/go-ycsb/pkg/generator"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"

	"github.com/mathieupost/jetflow/examples/simplebank/types"
	"github.com/mathieupost/jetflow/examples/simplebank/types/gen"
)

func main() {
	consumersAmount, err := strconv.Atoi(os.Getenv("CONSUMERS"))
	if err != nil {
		panic(errors.Wrap(err, "parsing CONSUMERS"))
	}
	natsHost := os.Getenv("NATS_HOST")
	if natsHost == "" {
		natsHost = "nats"
	}
	jaegerHost := os.Getenv("JAEGER_HOST")
	if jaegerHost == "" {
		jaegerHost = "jaeger"
	}

	tp, shutdown, err := tracing.NewProvider(jaegerHost+":4318", "api")
	if err != nil {
		log.Fatal("new tracing provider", err.Error())
	}
	defer shutdown()
	otel.SetTracerProvider(tp)
	defer func() {
		tp.ForceFlush(context.Background())
		println("flushed")
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nc, err := nats.Connect(natsHost)
	if err != nil {
		log.Fatal("connecting to NATS", err.Error())
	}

	js, err := natsjetstream.New(nc)
	if err != nil {
		log.Fatal("initializing JetStream instance", err.Error())
	}

	factoryMapping := gen.ProxyFactoryMapping()
	publisher := jetstream.NewPublisher(ctx, js, consumersAmount)
	client := jetflow.NewClient(factoryMapping, publisher)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RedirectSlashes)

	rs := rand.New(rand.NewSource(time.Now().Unix()))
	zipfGen := generator.NewZipfianWithItems(10000000, generator.ZipfianConstant)
	r.Get("/bench/{read}/{write}/{transact}", func(w http.ResponseWriter, r *http.Request) {
		readOps, _ := strconv.Atoi(chi.URLParam(r, "read"))
		writeOps, _ := strconv.Atoi(chi.URLParam(r, "write"))
		transactOps, _ := strconv.Atoi(chi.URLParam(r, "transact"))

		// Get a random user
		id1 := strconv.Itoa(int(zipfGen.Next(rs)))
		var user1 types.User
		client.Find(r.Context(), id1, &user1)

		// Determine the transaction
		total := readOps + writeOps + transactOps
		pick := rand.Intn(total)

		action := ""
		var err error
		switch {
		case pick < readOps:
			action = "read"
			_, err = user1.GetBalance(r.Context())
		case pick < writeOps+readOps:
			action = "write"
			_, err = user1.AddBalance(r.Context(), 1)
		default:
			action = "transact"
			id2 := strconv.Itoa(int(zipfGen.Next(rs)))
			for id1 == id2 {
				id2 = strconv.Itoa(int(zipfGen.Next(rs)))
			}

			var user2 types.User
			client.Find(r.Context(), id2, &user2)
			id1 += (" " + id2)

			_, _, err = user1.TransferBalance(r.Context(), user2, 1)
		}

		errStr := ""
		if err != nil {
			errStr = err.Error()
			render.Status(r, http.StatusInternalServerError)
			log.Println(action, id1, errStr)
		} else {
			log.Println(action, id1)
		}
		fmt.Fprintf(w, `{
  %s: %d,
  error: %s
}
`, action, pick, errStr)
	})

	r.Route("/users/{user}", func(r chi.Router) {
		r.Use(Operator[types.User](client, "user"))

		r.Get("/GetBalance", func(w http.ResponseWriter, r *http.Request) {
			user := r.Context().Value("user").(types.User)
			balance, _ := user.GetBalance(r.Context())
			if err != nil {
				log.Fatal(r.URL.Path, err.Error())
			}
			fmt.Fprintf(w, "{balance:%d}", balance)
		})

		r.Route("/TransferBalance/{user2}/{amount}", func(r chi.Router) {
			r.Use(Operator[types.User](client, "user2"))
			r.Use(Integer("amount"))

			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				user := r.Context().Value("user").(types.User)
				user2 := r.Context().Value("user2").(types.User)
				amount := r.Context().Value("amount").(int)
				balance1, balance2, err := user.TransferBalance(r.Context(), user2, amount)
				if err != nil {
					log.Fatal(r.URL.Path, err.Error())
				}
				fmt.Fprintf(w, "{balance1:%d,balance2:%d}", balance1, balance2)
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
