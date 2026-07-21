package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/myntra/bharat-ai-backend/internal/dto"
	"github.com/myntra/bharat-ai-backend/internal/services"
	"github.com/myntra/bharat-ai-backend/internal/utils"
	"github.com/myntra/bharat-ai-backend/internal/validator"
)

// ProductHandler serves the catalogue.
type ProductHandler struct {
	products *services.ProductService
	val      *validator.Validator
}

// NewProductHandler constructs a ProductHandler.
func NewProductHandler(products *services.ProductService, val *validator.Validator) *ProductHandler {
	return &ProductHandler{products: products, val: val}
}

// List handles GET /api/v1/products.
//
//	@Summary		List and filter products
//	@Tags			products
//	@Produce		json
//	@Param			q			query		string	false	"Free-text filter"
//	@Param			category	query		string	false	"Category name"
//	@Param			brand		query		string	false	"Brand name"
//	@Param			gender		query		string	false	"Women | Men | Unisex"
//	@Param			occasion	query		string	false	"Occasion"
//	@Param			style		query		string	false	"Style tag"
//	@Param			color		query		string	false	"Color"
//	@Param			minPrice	query		int		false	"Minimum price in rupees"
//	@Param			maxPrice	query		int		false	"Maximum price in rupees"
//	@Param			minRating	query		number	false	"Minimum rating"
//	@Param			sort		query		string	false	"relevance | price_asc | price_desc | rating | discount | newest"
//	@Param			page		query		int		false	"Page number"
//	@Param			pageSize	query		int		false	"Page size"
//	@Success		200			{object}	utils.Response{data=[]dto.Product,meta=utils.PageMeta}
//	@Router			/products [get]
func (h *ProductHandler) List(c *gin.Context) {
	var q dto.ProductQuery
	if !bindQuery(c, h.val, &q) {
		return
	}

	products, meta, err := h.products.List(c.Request.Context(), q)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OKWithMeta(c, products, meta)
}

// Search handles GET /api/v1/products/search.
//
// This is the same recommendation engine the assistant uses, entered through a
// structured query instead of a sentence — see internal/dto.ProductQuery.ToIntent.
// A shopper typing into the search bar and one asking the assistant the
// equivalent question get the same ranking, not two competing opinions about
// what "best" means.
//
//	@Summary		Rank products against a structured query using the recommendation engine
//	@Tags			products
//	@Produce		json
//	@Param			q			query		string	false	"Free-text query"
//	@Param			category	query		string	false	"Category name"
//	@Param			brand		query		string	false	"Brand name"
//	@Param			gender		query		string	false	"Women | Men | Unisex"
//	@Param			occasion	query		string	false	"Occasion"
//	@Param			style		query		string	false	"Style tag"
//	@Param			minPrice	query		int		false	"Minimum price in rupees"
//	@Param			maxPrice	query		int		false	"Maximum price in rupees"
//	@Param			sort		query		string	false	"relevance | price_asc | price_desc | rating | discount | newest"
//	@Success		200			{object}	utils.Response{data=[]dto.Product,meta=dto.SearchMeta}
//	@Router			/products/search [get]
func (h *ProductHandler) Search(c *gin.Context) {
	var q dto.ProductQuery
	if !bindQuery(c, h.val, &q) {
		return
	}
	q.Normalize()

	products, result, err := h.products.Search(c.Request.Context(), q)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OKWithMeta(c, products, dto.NewSearchMeta(q, result))
}

// GetByID handles GET /api/v1/products/:id.
//
// The path parameter accepts either a UUID or the catalogue slug
// ("prod-001") — see services.ProductService.GetByID.
//
//	@Summary		Get one product by ID or slug
//	@Tags			products
//	@Produce		json
//	@Param			id	path		string	true	"Product UUID or slug"
//	@Success		200	{object}	utils.Response{data=dto.Product}
//	@Failure		404	{object}	utils.ErrorResponse
//	@Router			/products/{id} [get]
func (h *ProductHandler) GetByID(c *gin.Context) {
	product, err := h.products.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, product)
}

// ByCategory handles GET /api/v1/products/category/:name.
//
//	@Summary		List products in one garment category
//	@Tags			products
//	@Produce		json
//	@Param			name		path		string	true	"Category name, e.g. 'Kurta Sets'"
//	@Param			page		query		int		false	"Page number"
//	@Param			pageSize	query		int		false	"Page size"
//	@Success		200			{object}	utils.Response{data=[]dto.Product,meta=utils.PageMeta}
//	@Failure		422			{object}	utils.ErrorResponse
//	@Router			/products/category/{name} [get]
func (h *ProductHandler) ByCategory(c *gin.Context) {
	var q dto.ProductQuery
	if !bindQuery(c, h.val, &q) {
		return
	}

	products, meta, err := h.products.ByCategory(c.Request.Context(), c.Param("name"), q)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OKWithMeta(c, products, meta)
}

// ByBrand handles GET /api/v1/products/brand/:name.
//
//	@Summary		List products from one brand
//	@Tags			products
//	@Produce		json
//	@Param			name		path		string	true	"Brand name"
//	@Param			page		query		int		false	"Page number"
//	@Param			pageSize	query		int		false	"Page size"
//	@Success		200			{object}	utils.Response{data=[]dto.Product,meta=utils.PageMeta}
//	@Router			/products/brand/{name} [get]
func (h *ProductHandler) ByBrand(c *gin.Context) {
	var q dto.ProductQuery
	if !bindQuery(c, h.val, &q) {
		return
	}

	products, meta, err := h.products.ByBrand(c.Request.Context(), c.Param("name"), q)
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OKWithMeta(c, products, meta)
}

// Categories handles GET /api/v1/categories.
//
//	@Summary		List garment categories with product counts
//	@Tags			products
//	@Produce		json
//	@Success		200	{object}	utils.Response{data=[]dto.Category}
//	@Router			/categories [get]
func (h *ProductHandler) Categories(c *gin.Context) {
	categories, err := h.products.Categories(c.Request.Context())
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, categories)
}

// Brands handles GET /api/v1/brands.
//
//	@Summary		List distinct brands in the catalogue
//	@Tags			products
//	@Produce		json
//	@Success		200	{object}	utils.Response{data=[]string}
//	@Router			/brands [get]
func (h *ProductHandler) Brands(c *gin.Context) {
	brands, err := h.products.Brands(c.Request.Context())
	if err != nil {
		utils.Fail(c, err)
		return
	}
	utils.OK(c, brands)
}
