// Copyright © 2019 Arrikto Inc.  All Rights Reserved.

package main

import (
	"context"
	"fmt"
	"github.com/coreos/go-oidc"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/myeong01/ai-playground/pkg/authz/authorizer"
	log "github.com/sirupsen/logrus"
	"github.com/tevino/abool"
	"github.com/wader/gormstore/v2"
	"golang.org/x/oauth2"
	"gorm.io/driver/sqlite" // Sqlite driver based on GGO
	"gorm.io/gorm"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Option Defaults
const (
	defaultHealthServerPort  = "8081"
	defaultServerHostname    = ""
	defaultServerPort        = "8080"
	defaultUserIDHeader      = "kubeflow-userid"
	defaultUserIDTokenHeader = "kubeflow-userid-token"
	defaultUserIDPrefix      = ""
	defaultUserIDClaim       = "email"
	defaultSessionMaxAge     = "86400"
)

// Issue: https://github.com/gorilla/sessions/issues/200
const secureCookieKeyPair = "notNeededBecauseCookieValueIsRandom"

type server struct {
	provider             *oidc.Provider
	oauth2Config         *oauth2.Config
	store                sessions.Store
	authorizer           *authorizer.Authorizer
	staticDestination    string
	sessionMaxAgeSeconds int
	apiHostName          string
	mainDomain           string
	userIDOpts
}

type userIDOpts struct {
	header      string
	tokenHeader string
	prefix      string
	claim       string
}

func main() {

	// Start readiness probe immediately
	log.Infof("Starting readiness probe at %v", defaultHealthServerPort)
	isReady := abool.New()
	go func() {
		log.Fatal(http.ListenAndServe(":"+defaultHealthServerPort, http.HandlerFunc(readiness(isReady))))
	}()

	/////////////
	// Options //
	/////////////

	// OIDC Provider
	providerURL := getURLEnvOrDie("OIDC_PROVIDER")
	authURL := os.Getenv("OIDC_AUTH_URL")
	// OIDC Client
	oidcScopes := clean(strings.Split(getEnvOrDie("OIDC_SCOPES"), " "))
	clientID := getEnvOrDie("CLIENT_ID")
	clientSecret := getEnvOrDie("CLIENT_SECRET")
	redirectURL := getURLEnvOrDie("REDIRECT_URL")
	staticDestination := os.Getenv("STATIC_DESTINATION_URL")
	whitelist := clean(strings.Split(os.Getenv("SKIP_AUTH_URI"), " "))
	// UserID Options
	userIDHeader := getEnvOrDefault("USERID_HEADER", defaultUserIDHeader)
	userIDTokenHeader := getEnvOrDefault("USERID_TOKEN_HEADER", defaultUserIDTokenHeader)
	userIDPrefix := getEnvOrDefault("USERID_PREFIX", defaultUserIDPrefix)
	userIDClaim := getEnvOrDefault("USERID_CLAIM", defaultUserIDClaim)
	// Server
	hostname := getEnvOrDefault("SERVER_HOSTNAME", defaultServerHostname)
	port := getEnvOrDefault("SERVER_PORT", defaultServerPort)
	// Store
	storePath := getEnvOrDie("STORE_PATH")
	// Sessions
	sessionMaxAge := getEnvOrDefault("SESSION_MAX_AGE", defaultSessionMaxAge)
	apiHostName := getEnvOrDefault("API_HOSTNAME", "")
	mainDomain := getEnvOrDefault("MAIN_DOMAIN", "")

	/////////////////////////////////////////////////////
	// Start server immediately for whitelisted routes //
	/////////////////////////////////////////////////////

	s := &server{}

	// Register handlers for routes
	router := mux.NewRouter()
	router.HandleFunc("/login/oidc", s.callback).Methods(http.MethodGet)
	router.HandleFunc("/logout", s.logout).Methods(http.MethodGet)
	router.PathPrefix("/").HandlerFunc(s.authenticate)

	// Start server
	log.Infof("Starting web server at %v:%v", hostname, port)
	stopCh := make(chan struct{})
	go func(stopCh chan struct{}) {
		log.Fatal(http.ListenAndServe(hostname+":"+port, handlers.CORS()(whitelistMiddleware(whitelist, isReady)(router))))
		close(stopCh)
	}(stopCh)

	/////////////////////////////////
	// Resume setup asynchronously //
	/////////////////////////////////

	// OIDC Discovery
	var provider *oidc.Provider
	var err error
	for {
		provider, err = oidc.NewProvider(context.Background(), providerURL.String())
		if err == nil {
			break
		}
		log.Errorf("OIDC provider setup failed, retrying in 10 seconds: %v", err)
		time.Sleep(10 * time.Second)
	}

	endpoint := provider.Endpoint()
	if authURL != "" {
		endpoint.AuthURL = authURL
	}

	oidcScopes = append(oidcScopes, oidc.ScopeOpenID)

	// Setup Store
	db, err := gorm.Open(sqlite.Open(storePath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error opening store: %v", err)
	}
	store := gormstore.New(db, []byte("thisismysecretlalalalalalalal"))
	quit := make(chan struct{})
	go store.PeriodicCleanup(1*time.Minute, quit)

	// Session Max-Age in seconds
	sessionMaxAgeSeconds, err := strconv.Atoi(sessionMaxAge)
	if err != nil {
		log.Fatalf("Couldn't convert session MaxAge to int: %v", err)
	}

	// Set the server values.
	// The isReady atomic variable should protect it from concurrency issues.

	*s = server{
		provider: provider,
		oauth2Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     endpoint,
			RedirectURL:  redirectURL.String(),
			Scopes:       oidcScopes,
		},
		// TODO: Add support for Redis
		store:             store,
		staticDestination: staticDestination,
		authorizer:        authorizer.New(),
		userIDOpts: userIDOpts{
			header:      userIDHeader,
			tokenHeader: userIDTokenHeader,
			prefix:      userIDPrefix,
			claim:       userIDClaim,
		},
		sessionMaxAgeSeconds: sessionMaxAgeSeconds,
		apiHostName:          apiHostName,
		mainDomain:           mainDomain,
	}

	fmt.Println("apiHostName:", apiHostName)
	fmt.Println("mainDomain:", mainDomain)

	// Setup complete, mark server ready
	isReady.Set()

	// Block until server exits
	<-stopCh
}
