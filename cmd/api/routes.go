package main

import (
	"expvar"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   app.config.cors.trustedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	//Use alice to make a global middleware chain.
	globalMiddleware := alice.New(app.metrics, app.recoverPanic, app.authenticate).Then
	// Dynamic Middleware, these will apply to only select routes
	dynamicMiddleware := alice.New(app.requireAuthenticatedUser, app.requireActivatedUser)

	// Apply the global middleware to the router
	router.Use(globalMiddleware)

	// Make our categorized routes
	v1Router := chi.NewRouter()

	v1Router.Mount("/", app.generalRoutes())
	v1Router.Mount("/api", app.userRoutes())

	// MUsic
	v1Router.Mount("/musical", app.musicalRoutes(&dynamicMiddleware))

	// Moount the v1Router to the main base router
	router.Mount("/v1", v1Router)
	return router
}

// generalRoutes() provides a router for the general routes.
// Mounted rirectly after our version url. They contaon sanity and
// health checks. Probably add other AOB's here.
func (app *application) generalRoutes() chi.Router {
	generalRoutes := chi.NewRouter()
	// /debug/vars : for expvar, wrapping it in a handler func for assertion otherwise it complains
	generalRoutes.Get("/debug/vars", func(w http.ResponseWriter, r *http.Request) {
		expvar.Handler().ServeHTTP(w, r)
	})
	generalRoutes.Get("/health", app.healthcheckHandler)
	return generalRoutes
}

// userRoutes() is a method that returns a chi.Router that contains all the routes for the users
func (app *application) userRoutes() chi.Router {
	userRoutes := chi.NewRouter()
	userRoutes.Post("/", app.registerUserHandler)
	userRoutes.Post("/authentication", app.createAuthenticationApiKeyHandler)
	// /activation : for activating accounts
	userRoutes.Put("/activated", app.activateUserHandler)
	return userRoutes
}

func (app *application) musicalRoutes(dynamicMiddleware *alice.Chain) chi.Router {
	musicalRoutes := chi.NewRouter()
	// /musicalnews : for fetching all musical news
	musicalRoutes.With(dynamicMiddleware.Then).Get("/news", app.getAllMusicalNews)
	// /trends : for fetching music trends from Last.fm
	musicalRoutes.With(dynamicMiddleware.Then).Get("/trends", app.getAllMusicTrends)
	// /lyrics : for fetching song lyrics
	musicalRoutes.With(dynamicMiddleware.Then).Get("/lyrics", app.getLyrics)
	return musicalRoutes

}
