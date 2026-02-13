package handlers

import (
	"net/http"
	"strconv"

	"autoimport-kz/internal/middleware"
	"autoimport-kz/internal/models"
	"autoimport-kz/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
)

type PurchaseHandler struct {
	purchases   *services.PurchaseService
	renderer    *TemplateRenderer
	store       *sessions.CookieStore
	sessionName string
}

func NewPurchaseHandler(purchases *services.PurchaseService, renderer *TemplateRenderer, store *sessions.CookieStore, sessionName string) *PurchaseHandler {
	return &PurchaseHandler{
		purchases:   purchases,
		renderer:    renderer,
		store:       store,
		sessionName: sessionName,
	}
}

type purchasesPageData struct {
	Title           string
	ContentTemplate string
	Items           []models.Purchase
	Flash           string
}

type purchaseDetailPageData struct {
	Title           string
	ContentTemplate string
	Item            models.Purchase
	User            models.User
	Flash           string
}

func (h *PurchaseHandler) Buy(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	carID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || carID <= 0 {
		http.Error(w, "invalid car id", http.StatusBadRequest)
		return
	}

	if _, err := h.purchases.Buy(r.Context(), user.ID, carID); err != nil {
		http.Error(w, "failed to create purchase", http.StatusInternalServerError)
		return
	}

	setFlash(w, r, h.store, h.sessionName, "Car purchased successfully.")
	http.Redirect(w, r, "/purchases", http.StatusFound)
}

func (h *PurchaseHandler) MyPurchases(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	items, err := h.purchases.MyPurchases(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "failed to load purchases", http.StatusInternalServerError)
		return
	}

	h.renderer.Render(w, "purchases.html", purchasesPageData{
		Title:           "My Purchases",
		ContentTemplate: "purchases_content",
		Items:           items,
		Flash:           popFlash(w, r, h.store, h.sessionName),
	})
}

func (h *PurchaseHandler) Detail(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	purchaseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || purchaseID <= 0 {
		http.NotFound(w, r)
		return
	}

	item, err := h.purchases.GetByID(r.Context(), purchaseID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if user.Role != "admin" && item.UserID != user.ID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	h.renderer.Render(w, "purchase_detail.html", purchaseDetailPageData{
		Title:           "Purchase Details",
		ContentTemplate: "purchase_detail_content",
		Item:            item,
		User:            user,
		Flash:           popFlash(w, r, h.store, h.sessionName),
	})
}

func (h *PurchaseHandler) MarkPaid(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	purchaseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || purchaseID <= 0 {
		http.Error(w, "invalid purchase id", http.StatusBadRequest)
		return
	}

	if err := h.purchases.MarkPaid(r.Context(), user.ID, purchaseID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	setFlash(w, r, h.store, h.sessionName, "Purchase marked as paid.")
	http.Redirect(w, r, "/purchases/"+strconv.FormatInt(purchaseID, 10), http.StatusFound)
}

func (h *PurchaseHandler) MarkDelivered(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if user.Role != "admin" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	purchaseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || purchaseID <= 0 {
		http.Error(w, "invalid purchase id", http.StatusBadRequest)
		return
	}

	if err := h.purchases.MarkDelivered(r.Context(), purchaseID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	setFlash(w, r, h.store, h.sessionName, "Purchase marked as delivered.")
	http.Redirect(w, r, "/purchases/"+strconv.FormatInt(purchaseID, 10), http.StatusFound)
}
