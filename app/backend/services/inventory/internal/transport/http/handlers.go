package http

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/embedded-market/backend/pkg/apperr"
	"github.com/embedded-market/backend/pkg/httpx"
	"github.com/embedded-market/backend/pkg/middleware"
	pkgvalidator "github.com/embedded-market/backend/pkg/validator"
	"github.com/embedded-market/backend/services/inventory/internal/application"
	"github.com/embedded-market/backend/services/inventory/internal/domain"
)

type UseCases struct { Inventory *application.InventoryUseCase }

type Handler struct { uc UseCases }
func NewHandler(uc UseCases) *Handler { return &Handler{uc: uc} }

func (h *Handler) ListWarehouses(c *gin.Context) {
	items, err := h.uc.Inventory.ListWarehouses(c.Request.Context())
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, gin.H{"items": items})
}

func (h *Handler) CreateWarehouse(c *gin.Context) {
	var req warehouseCreateRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	item, err := h.uc.Inventory.CreateWarehouse(c.Request.Context(), application.CreateWarehouseInput{Code: req.Code, Name: req.Name, Location: req.Location})
	if err != nil { httpx.Fail(c, err); return }
	httpx.Created(c, item)
}

func (h *Handler) ListStock(c *gin.Context) {
	page := httpx.ParsePageParams(c)
	filters := domain.StockFilters{}
	if raw := strings.TrimSpace(c.Query("product_id")); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil { httpx.Fail(c, apperr.Invalid("invalid_product_id", "the product id is not a valid UUID")); return }
		filters.ProductID = &id
	}
	if raw := strings.TrimSpace(c.Query("warehouse_id")); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil { httpx.Fail(c, apperr.Invalid("invalid_warehouse_id", "the warehouse id is not a valid UUID")); return }
		filters.WarehouseID = &id
	}
	items, total, err := h.uc.Inventory.ListStock(c.Request.Context(), filters, page.Limit(), page.Offset())
	if err != nil { httpx.Fail(c, err); return }
	httpx.Page(c, items, page.BuildPagination(total))
}

func (h *Handler) AdjustStock(c *gin.Context) {
	var req adjustStockRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	productID, err := uuid.Parse(req.ProductID)
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_product_id", "the product id is not a valid UUID")); return }
	warehouseID, err := uuid.Parse(req.WarehouseID)
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_warehouse_id", "the warehouse id is not a valid UUID")); return }
	item, err := h.uc.Inventory.AdjustStock(c.Request.Context(), application.AdjustStockInput{ProductID: productID, WarehouseID: warehouseID, Delta: req.Delta}, requestContext(c))
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, item)
}

func (h *Handler) ReserveStock(c *gin.Context) {
	var req reserveStockRequest
	if err := bindJSON(c, &req); err != nil { httpx.Fail(c, err); return }
	productID, err := uuid.Parse(req.ProductID)
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_product_id", "the product id is not a valid UUID")); return }
	warehouseID, err := uuid.Parse(req.WarehouseID)
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_warehouse_id", "the warehouse id is not a valid UUID")); return }
	item, err := h.uc.Inventory.ReserveStock(c.Request.Context(), application.ReserveStockInput{ProductID: productID, WarehouseID: warehouseID, Reference: req.Reference, Quantity: req.Quantity}, requestContext(c))
	if err != nil { httpx.Fail(c, err); return }
	httpx.Created(c, item)
}

func (h *Handler) ReleaseReservation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil { httpx.Fail(c, apperr.Invalid("invalid_reservation_id", "the reservation id is not a valid UUID")); return }
	item, err := h.uc.Inventory.ReleaseReservation(c.Request.Context(), id, requestContext(c))
	if err != nil { httpx.Fail(c, err); return }
	httpx.OK(c, item)
}

func (h *Handler) ListReservations(c *gin.Context) {
	page := httpx.ParsePageParams(c)
	items, total, err := h.uc.Inventory.ListReservations(c.Request.Context(), page.Limit(), page.Offset())
	if err != nil { httpx.Fail(c, err); return }
	httpx.Page(c, items, page.BuildPagination(total))
}

func bindJSON(c *gin.Context, target any) error {
	if err := c.ShouldBindJSON(target); err != nil { return apperr.Invalid("invalid_json", "the request body is invalid").WithCause(err) }
	if err := pkgvalidator.Struct(target); err != nil { return err }
	return nil
}

func requestContext(c *gin.Context) application.RequestContext {
	var accountID uuid.UUID
	if raw := middleware.AccountID(c); raw != "" { accountID, _ = uuid.Parse(raw) }
	return application.RequestContext{CorrelationID: middleware.CorrelationID(c), AccountID: accountID}
}
