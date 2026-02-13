package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"autoimport-kz/internal/config"
	"autoimport-kz/internal/handlers"
	appmw "autoimport-kz/internal/middleware"
	"autoimport-kz/internal/repositories"
	"autoimport-kz/internal/services"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := config.Load()

	dbConn, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	if err := dbConn.Ping(); err != nil {
		log.Fatal(err)
	}

	repoDB := repositories.NewDB(dbConn)
	userRepo := repositories.NewUserRepo(repoDB)
	countryRepo := repositories.NewCountryRepo(repoDB)
	carRepo := repositories.NewCarRepo(repoDB)
	calcRepo := repositories.NewCalculationRepo(repoDB)
	purchaseRepo := repositories.NewPurchaseRepo(repoDB)

	authService := services.NewAuthService(userRepo)
	countryService := services.NewCountryService(countryRepo)
	carService := services.NewCarService(carRepo)
	calcService := services.NewCalculationService(carRepo, countryRepo, calcRepo)
	purchaseService := services.NewPurchaseService(carRepo, purchaseRepo)
	dashboardService := services.NewDashboardService(userRepo, carRepo, countryRepo, calcRepo, purchaseRepo)

	renderer, err := handlers.NewTemplateRenderer(cfg.TemplateDir)
	if err != nil {
		log.Fatal(err)
	}

	store := sessions.NewCookieStore([]byte(cfg.SessionKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   false,
	}

	if err := ensureDefaultAdmin(userRepo, cfg.DefaultAdmin, cfg.DefaultPass); err != nil {
		log.Fatal(err)
	}

	authHandler := handlers.NewAuthHandler(authService, renderer, store, cfg.SessionName)
	carHandler := handlers.NewCarHandler(carService, countryService, renderer, store, cfg.SessionName)
	countryHandler := handlers.NewCountryHandler(countryService, renderer, store, cfg.SessionName)
	calcHandler := handlers.NewCalculatorHandler(carService, countryService, calcService, renderer)
	purchaseHandler := handlers.NewPurchaseHandler(purchaseService, renderer, store, cfg.SessionName)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService, renderer, store, cfg.SessionName)

	r := chi.NewRouter()
	r.Use(chimw.RealIP)
	r.Use(chimw.RequestID)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(appmw.RequestLogger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		renderer.Render(w, "home.html", map[string]any{
			"Title":           "AutoImport KZ",
			"ContentTemplate": "home_content",
		})
	})

	r.Get("/register", authHandler.ShowRegister)
	r.Post("/auth/register", authHandler.Register)
	r.Get("/login", authHandler.ShowLogin)
	r.Post("/auth/login", authHandler.Login)
	r.Post("/auth/logout", authHandler.Logout)

	r.Group(func(auth chi.Router) {
		auth.Use(appmw.AuthMiddleware(store, cfg.SessionName, userRepo))
		auth.Get("/me", authHandler.Profile)
		auth.Get("/dashboard", dashboardHandler.Show)
		auth.Get("/calculator", calcHandler.ShowCalculator)
		auth.Post("/calculator/calculate", calcHandler.Calculate)
		auth.Get("/history", calcHandler.History)
		auth.Post("/cars/{id}/buy", purchaseHandler.Buy)
		auth.Get("/purchases", purchaseHandler.MyPurchases)
		auth.Get("/purchases/{id}", purchaseHandler.Detail)
		auth.Post("/purchases/{id}/pay", purchaseHandler.MarkPaid)
	})

	r.Get("/cars", carHandler.List)
	r.Get("/cars/{id}", carHandler.Detail)

	r.Route("/admin", func(admin chi.Router) {
		admin.Use(appmw.AuthMiddleware(store, cfg.SessionName, userRepo))
		admin.Use(appmw.RoleMiddleware("admin"))

		admin.Get("/cars", carHandler.AdminList)
		admin.Post("/cars", carHandler.AdminCreate)
		admin.Post("/cars/{id}/update", carHandler.AdminUpdateByPost)
		admin.Post("/cars/{id}/delete", carHandler.AdminDeleteByPost)
		admin.Put("/cars/{id}", carHandler.AdminUpdate)
		admin.Delete("/cars/{id}", carHandler.AdminDelete)

		admin.Get("/countries", countryHandler.AdminList)
		admin.Post("/countries", countryHandler.AdminCreate)
		admin.Post("/countries/{id}/update", countryHandler.AdminUpdateByPost)
		admin.Post("/countries/{id}/delete", countryHandler.AdminDeleteByPost)
		admin.Put("/countries/{id}", countryHandler.AdminUpdate)
		admin.Delete("/countries/{id}", countryHandler.AdminDelete)
		admin.Post("/purchases/{id}/deliver", purchaseHandler.MarkDelivered)
	})

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.StaticDir))))

	log.Printf("listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, r); err != nil {
		log.Fatal(err)
	}
}

func ensureDefaultAdmin(users *repositories.UserRepo, email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return users.UpsertAdmin(context.Background(), "Admin", email, string(hash))
}
