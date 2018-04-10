package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"bitbucket.org/Rusty1958/shakingdog/auth"
	"bitbucket.org/Rusty1958/shakingdog/config"
	"bitbucket.org/Rusty1958/shakingdog/db"
	"bitbucket.org/Rusty1958/shakingdog/webserver"

	"github.com/gorilla/mux"
)

var (
	confFile string
	// this doubles as global store for the things we
	// need to pass to the request handlers
	handlerContext *webserver.HandlerContext
)


func init() {
	flag.StringVar(&confFile, "f", "", "Path to the configuration file.")

	handlerContext = &webserver.HandlerContext{}
}

func main() {
	var (
		cfg *config.Config
		err error
	)

  // parse CLI arguments
	flag.Parse()
  if flag.NFlag() < 1 {
    fmt.Println("== Shaking Dog Register ==\n")
    flag.PrintDefaults()
    return
  }

	// read in the config file
  cfg, err = config.Load(confFile)
	if err != nil {
		log.Fatalf("Error reading configuration file - %v", err)
	}
	handlerContext.Config = cfg

	// read the CA file once instead of every request
	// get the lists of certs and keys
	cpaths := strings.Split(cfg.Server.CertPaths, ",")
	kpaths := strings.Split(cfg.Server.KeyPaths, ",")

	// create the server
	s, l, err := webserver.NewServer(
		cfg.Server.Addr,
		cfg.Server.CAPath,
		cpaths,
		kpaths,
	)
	if err != nil {
		log.Fatalf("Error creating web server - %v", err)
	}

	// Okta auth checker
	handlerContext.Okta = auth.NewOktaAuth(
		fmt.Sprintf("https://%s", cfg.Okta.Host),
		cfg.Okta.ClientID,
		cfg.Okta.ClientSecret,
		cfg.Server.BaseURL,
		fmt.Sprintf("https://%s%s%s",
			cfg.Server.PublicHost,
			cfg.Server.BaseURL,
			cfg.Okta.AuthPath,
	))

	// build routes
	s.Handler = BuildRouter(cfg, handlerContext.Okta)

	// create DB connection
	handlerContext.DBConnection, err = db.NewMySQLConn(
		cfg.Server.DBHost,
		cfg.Server.DBName,
		cfg.Server.DBUserName,
		cfg.Server.DBPassword,
	)
	if err != nil {
		log.Fatalf("Error establishing database connection - %v", err)
	}

	// start listening and wait for graceful shutdown
	// https://github.com/gorilla/mux#graceful-shutdown
	log.Printf("Starting web server - addr=%s", cfg.Server.Addr)
  go func() {
      if err := s.Serve(l); err != nil {
          log.Println(err)
      }
  }()
  c := make(chan os.Signal, 1)
  signal.Notify(c, os.Interrupt) // SIGINT (Ctrl+C)
  <-c
  ctx, cancel := context.WithTimeout(context.Background(), 15)
  defer cancel()
  s.Shutdown(ctx)
  log.Println("Web server gracefully shutdown. Bye bye!")
  os.Exit(0)
}

func BuildRouter(cfg *config.Config, oktaAuth *auth.Okta) http.Handler {
	router := mux.NewRouter()

	// "root" serves up static UI files
	router.PathPrefix("/app").Handler(
		http.StripPrefix(
			fmt.Sprintf("%s/app", cfg.Server.BaseURL),
			http.FileServer(http.Dir(cfg.Server.StaticPath))),
	)

	// multiple dogs fetch
	router.HandleFunc(
		fmt.Sprintf("%s/api/dogs", cfg.Server.BaseURL),
		HandlerWithContext(webserver.DogsHandler),
	)

	// single dog fetch
	router.HandleFunc(
		fmt.Sprintf("%s/api/dog/{id:[0-9]+}", cfg.Server.BaseURL),
		HandlerWithContext(webserver.DogHandler),
	)

	// shaking dog admin
/*	router.Handle(
		fmt.Sprintf("%s/api/admin/shakingdog", cfg.Server.BaseURL),
		oktaAuth.SecuredHandler(
			HandlerWithContext(webserver.ShakingDogHandler),
			HandlerWithContext(webserver.NeedAuthHandler),
	))*/

	// cecs admin
/*	router.Handle(
		fmt.Sprintf("%s/api/admin/cecs", cfg.Server.BaseURL),
		oktaAuth.SecuredHandler(
			HandlerWithContext(webserver.CecsHandler),
			HandlerWithContext(webserver.NeedAuthHandler),
	))*/

	// sets the state cookie and bounces user to the Okta login page
	router.Handle(
		fmt.Sprintf("%s%s", cfg.Server.BaseURL, cfg.Okta.LoginPath),
		oktaAuth.LoginHandler(),
	)

	// Okta sends us back here after auth
	router.Handle(
		fmt.Sprintf("%s%s", cfg.Server.BaseURL, cfg.Okta.AuthPath),
		oktaAuth.AuthCallbackHandler(),
	)

	// unmatched redirect to "/app"
	router.Handle(
		fmt.Sprintf("%s/", cfg.Server.BaseURL),
		http.RedirectHandler(
			fmt.Sprintf("%s/app", cfg.Server.BaseURL),
			http.StatusMovedPermanently,
	))

	return router
}

func HandlerWithContext(handler func(http.ResponseWriter, *http.Request, *webserver.HandlerContext)) (http.Handler) {
	return http.HandlerFunc(func(w http.ResponseWriter, q *http.Request) {
		// hand our context to the handler function
		handler(w, q, handlerContext)
	})
}
