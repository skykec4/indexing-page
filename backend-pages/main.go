package main

import (
	"fmt"
	"log"
	"net/http"
	"pages/internal/database"
	"pages/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

const PORT = 3000

func main() {
	// 데이터베이스 연결
	db, err := database.NewDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// 라우터 생성
	r := chi.NewRouter()

	// 미들웨어
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API 라우트
	r.Route("/api", func(r chi.Router) {
		h := handler.NewHandler(db)

		// 사이트 관련 라우트
		r.Route("/sites", func(r chi.Router) {
			r.Get("/", h.GetSites)
			r.Post("/", h.CreateSite)
			r.Route("/{siteCode}", func(r chi.Router) {
				r.Get("/menu", h.GetSiteMenu)
				r.Route("/groups", func(r chi.Router) {
					r.Get("/", h.GetPageGroups)
					r.Post("/", h.CreatePageGroup)
					r.Route("/{groupId}", func(r chi.Router) {
						r.Put("/", h.UpdatePageGroup)
						r.Delete("/", h.DeletePageGroup)
					})
				})
				r.Route("/pages", func(r chi.Router) {
					r.Get("/", h.ListPages)
					r.Post("/", h.CreatePage)
					r.Route("/{pageID}", func(r chi.Router) {
						r.Get("/", h.GetPage)
						r.Put("/", h.UpdatePage)
						r.Delete("/", h.DeletePage)
					})
				})
			})
		})
	})

	log.Printf("Server starting on port %d", PORT)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", PORT), r); err != nil {
		log.Fatal(err)
	}
}
