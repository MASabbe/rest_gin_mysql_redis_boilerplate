package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func main() {
	nameFlag := flag.String("name", "", "Feature name in lowercase singular (e.g. product, order)")
	flag.Parse()

	featureName := strings.ToLower(strings.TrimSpace(*nameFlag))
	if featureName == "" {
		fmt.Fprintln(os.Stderr, "Error: -name flag is required (e.g. go run ./cmd/scaffold -name=product)")
		os.Exit(1)
	}

	if err := validateFeatureName(featureName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := ScaffoldFeature(".", featureName); err != nil {
		fmt.Fprintf(os.Stderr, "Error scaffolding feature: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully scaffolded canonical feature '%s' under internal/modules/%s/\n", featureName, featureName)
}

func validateFeatureName(name string) error {
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return fmt.Errorf("feature name '%s' contains invalid characters (use only letters, digits, and underscores)", name)
		}
	}
	return nil
}

func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	var result string
	for _, p := range parts {
		if len(p) > 0 {
			result += strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return result
}

// ScaffoldFeature generates canonical Clean Architecture template files for the given feature name.
func ScaffoldFeature(baseDir, featureName string) error {
	pascalName := toPascalCase(featureName)
	moduleDir := filepath.Join(baseDir, "internal", "modules", featureName)

	files := map[string]string{
		// Domain Entity
		filepath.Join(moduleDir, "domain", "entity", fmt.Sprintf("%s.go", featureName)): fmt.Sprintf(`package entity

import (
	"strings"
	"time"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/google/uuid"
)

// %[1]s represents the %[1]s domain entity.
type %[1]s struct {
	ID        string    `+"`"+`json:"id"`+"`"+`
	Name      string    `+"`"+`json:"name"`+"`"+`
	CreatedAt time.Time `+"`"+`json:"created_at"`+"`"+`
	UpdatedAt time.Time `+"`"+`json:"updated_at"`+"`"+`
}

// New%[1]s creates a new validated %[1]s entity.
func New%[1]s(name string) (*%[1]s, error) {
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return nil, appErrors.NewValidationError("name is required", appErrors.FieldError{
			Field:   "name",
			Message: "cannot be empty",
		})
	}

	now := time.Now().UTC()
	return &%[1]s{
		ID:        uuid.New().String(),
		Name:      cleanName,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
`, pascalName),

		// Domain Repository Interface
		filepath.Join(moduleDir, "domain", "repository", fmt.Sprintf("%s_repository.go", featureName)): fmt.Sprintf(`package repository

import (
	"context"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/%[2]s/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
)

// %[1]sRepository defines persistence operations for %[1]s.
type %[1]sRepository interface {
	Create(ctx context.Context, item *entity.%[1]s) error
	FindByID(ctx context.Context, id string) (*entity.%[1]s, error)
	List(ctx context.Context, p pagination.Pagination) ([]*entity.%[1]s, int64, error)
	Update(ctx context.Context, item *entity.%[1]s) error
	Delete(ctx context.Context, id string) error
}
`, pascalName, featureName),

		// Application Commands
		filepath.Join(moduleDir, "application", "command", fmt.Sprintf("%s_commands.go", featureName)): fmt.Sprintf(`package command

import (
	"strings"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
)

type Create%[1]sCommand struct {
	Name string `+"`"+`json:"name"`+"`"+`
}

func (c Create%[1]sCommand) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return appErrors.NewValidationError("name is required", appErrors.FieldError{
			Field:   "name",
			Message: "cannot be empty",
		})
	}
	return nil
}

type Update%[1]sCommand struct {
	ID   string `+"`"+`json:"id"`+"`"+`
	Name string `+"`"+`json:"name"`+"`"+`
}

func (c Update%[1]sCommand) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return appErrors.NewValidationError("ID is required", appErrors.FieldError{
			Field:   "id",
			Message: "cannot be empty",
		})
	}
	return nil
}

type Delete%[1]sCommand struct {
	ID string `+"`"+`json:"id"`+"`"+`
}

func (c Delete%[1]sCommand) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return appErrors.NewValidationError("ID is required", appErrors.FieldError{
			Field:   "id",
			Message: "cannot be empty",
		})
	}
	return nil
}
`, pascalName),

		// Application Queries
		filepath.Join(moduleDir, "application", "query", fmt.Sprintf("%s_queries.go", featureName)): fmt.Sprintf(`package query

import (
	"strings"

	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
)

type Get%[1]sByIDQuery struct {
	ID string `+"`"+`json:"id"`+"`"+`
}

func (q Get%[1]sByIDQuery) Validate() error {
	if strings.TrimSpace(q.ID) == "" {
		return appErrors.NewValidationError("ID is required", appErrors.FieldError{
			Field:   "id",
			Message: "cannot be empty",
		})
	}
	return nil
}

type List%[1]sQuery struct {
	Pagination pagination.Pagination
}
`, pascalName),

		// Application DTOs
		filepath.Join(moduleDir, "application", "dto", fmt.Sprintf("%s_dto.go", featureName)): fmt.Sprintf(`package dto

import (
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/%[2]s/domain/entity"
)

type %[1]sDTO struct {
	ID        string    `+"`"+`json:"id"`+"`"+`
	Name      string    `+"`"+`json:"name"`+"`"+`
	CreatedAt time.Time `+"`"+`json:"created_at"`+"`"+`
	UpdatedAt time.Time `+"`"+`json:"updated_at"`+"`"+`
}

func To%[1]sDTO(item *entity.%[1]s) %[1]sDTO {
	if item == nil {
		return %[1]sDTO{}
	}
	return %[1]sDTO{
		ID:        item.ID,
		Name:      item.Name,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func To%[1]sDTOList(items []*entity.%[1]s) []%[1]sDTO {
	list := make([]%[1]sDTO, 0, len(items))
	for _, item := range items {
		list = append(list, To%[1]sDTO(item))
	}
	return list
}
`, pascalName, featureName),

		// Application Service
		filepath.Join(moduleDir, "application", "service.go"): fmt.Sprintf(`package application

import (
	"context"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/%[2]s/application/command"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/%[2]s/application/dto"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/%[2]s/application/query"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/%[2]s/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/%[2]s/domain/repository"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
)

type %[1]sService interface {
	Create%[1]s(ctx context.Context, cmd command.Create%[1]sCommand) (*dto.%[1]sDTO, error)
	Get%[1]s(ctx context.Context, q query.Get%[1]sByIDQuery) (*dto.%[1]sDTO, error)
	List%[1]s(ctx context.Context, q query.List%[1]sQuery) ([]dto.%[1]sDTO, pagination.Meta, error)
	Update%[1]s(ctx context.Context, cmd command.Update%[1]sCommand) (*dto.%[1]sDTO, error)
	Delete%[1]s(ctx context.Context, cmd command.Delete%[1]sCommand) error
}

type %[2]sService struct {
	repo repository.%[1]sRepository
}

func New%[1]sService(repo repository.%[1]sRepository) %[1]sService {
	return &%[2]sService{repo: repo}
}

func (s *%[2]sService) Create%[1]s(ctx context.Context, cmd command.Create%[1]sCommand) (*dto.%[1]sDTO, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	item, err := entity.New%[1]s(cmd.Name)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}

	res := dto.To%[1]sDTO(item)
	return &res, nil
}

func (s *%[2]sService) Get%[1]s(ctx context.Context, q query.Get%[1]sByIDQuery) (*dto.%[1]sDTO, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}

	item, err := s.repo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	res := dto.To%[1]sDTO(item)
	return &res, nil
}

func (s *%[2]sService) List%[1]s(ctx context.Context, q query.List%[1]sQuery) ([]dto.%[1]sDTO, pagination.Meta, error) {
	items, total, err := s.repo.List(ctx, q.Pagination)
	if err != nil {
		return nil, pagination.Meta{}, err
	}

	meta := pagination.NewMeta(q.Pagination.Page, q.Pagination.PageSize, total)
	return dto.To%[1]sDTOList(items), meta, nil
}

func (s *%[2]sService) Update%[1]s(ctx context.Context, cmd command.Update%[1]sCommand) (*dto.%[1]sDTO, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	item, err := s.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	item.Name = cmd.Name
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}

	res := dto.To%[1]sDTO(item)
	return &res, nil
}

func (s *%[2]sService) Delete%[1]s(ctx context.Context, cmd command.Delete%[1]sCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	return s.repo.Delete(ctx, cmd.ID)
}
`, pascalName, featureName),

		// Infrastructure In-Memory Repository
		filepath.Join(moduleDir, "infrastructure", "persistence", fmt.Sprintf("inmemory_%s_repository.go", featureName)): fmt.Sprintf(`package persistence

import (
	"context"
	"sync"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/%[2]s/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/%[2]s/domain/repository"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
)

type InMemory%[1]sRepository struct {
	mu    sync.RWMutex
	items map[string]*entity.%[1]s
}

func NewInMemory%[1]sRepository() *InMemory%[1]sRepository {
	return &InMemory%[1]sRepository{
		items: make(map[string]*entity.%[1]s),
	}
}

func (r *InMemory%[1]sRepository) Create(ctx context.Context, item *entity.%[1]s) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[item.ID] = item
	return nil
}

func (r *InMemory%[1]sRepository) FindByID(ctx context.Context, id string) (*entity.%[1]s, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[id]
	if !ok {
		return nil, appErrors.NewNotFoundError("%[2]s not found")
	}
	return item, nil
}

func (r *InMemory%[1]sRepository) List(ctx context.Context, p pagination.Pagination) ([]*entity.%[1]s, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []*entity.%[1]s
	for _, item := range r.items {
		all = append(all, item)
	}

	total := int64(len(all))
	offset := p.Offset()
	limit := p.Limit()

	if offset >= len(all) {
		return []*entity.%[1]s{}, total, nil
	}

	end := offset + limit
	if end > len(all) {
		end = len(all)
	}

	return all[offset:end], total, nil
}

func (r *InMemory%[1]sRepository) Update(ctx context.Context, item *entity.%[1]s) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[item.ID] = item
	return nil
}

func (r *InMemory%[1]sRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, id)
	return nil
}

var _ repository.%[1]sRepository = (*InMemory%[1]sRepository)(nil)
`, pascalName, featureName),

		// Delivery Request
		filepath.Join(moduleDir, "delivery", "http", "request", fmt.Sprintf("%s_request.go", featureName)): fmt.Sprintf(`package request

type Create%[1]sRequest struct {
	Name string `+"`"+`json:"name" binding:"required,min=2,max=255"`+"`"+`
}

type Update%[1]sRequest struct {
	Name string `+"`"+`json:"name" binding:"required,min=2,max=255"`+"`"+`
}
`, pascalName),

		// Delivery Response
		filepath.Join(moduleDir, "delivery", "http", "response", fmt.Sprintf("%s_response.go", featureName)): fmt.Sprintf(`package response

import (
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/%[2]s/application/dto"
)

type %[1]sResponse struct {
	ID        string    `+"`"+`json:"id"`+"`"+`
	Name      string    `+"`"+`json:"name"`+"`"+`
	CreatedAt time.Time `+"`"+`json:"created_at"`+"`"+`
	UpdatedAt time.Time `+"`"+`json:"updated_at"`+"`"+`
}

func From%[1]sDTO(d *dto.%[1]sDTO) %[1]sResponse {
	if d == nil {
		return %[1]sResponse{}
	}
	return %[1]sResponse{
		ID:        d.ID,
		Name:      d.Name,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func From%[1]sDTOList(dtos []dto.%[1]sDTO) []%[1]sResponse {
	list := make([]%[1]sResponse, 0, len(dtos))
	for _, d := range dtos {
		list = append(list, From%[1]sDTO(&d))
	}
	return list
}
`, pascalName, featureName),

		// Delivery Handler
		filepath.Join(moduleDir, "delivery", "http", "handler", fmt.Sprintf("%s_handler.go", featureName)): fmt.Sprintf(`package handler

import (
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/%[2]s/application"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/%[2]s/application/command"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/%[2]s/application/query"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/%[2]s/delivery/http/request"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/%[2]s/delivery/http/response"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
	sharedResponse "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type %[1]sHandler struct {
	service application.%[1]sService
}

func New%[1]sHandler(service application.%[1]sService) *%[1]sHandler {
	return &%[1]sHandler{service: service}
}

func (h *%[1]sHandler) Create(c *gin.Context) {
	var req request.Create%[1]sRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedResponse.Error(c, appErrors.NewValidationError("invalid request payload", appErrors.FieldError{
			Field:   "body",
			Message: err.Error(),
		}))
		return
	}

	result, err := h.service.Create%[1]s(c.Request.Context(), command.Create%[1]sCommand{Name: req.Name})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.Created(c, "%[1]s created successfully", response.From%[1]sDTO(result))
}

func (h *%[1]sHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	result, err := h.service.Get%[1]s(c.Request.Context(), query.Get%[1]sByIDQuery{ID: id})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "%[1]s retrieved successfully", response.From%[1]sDTO(result))
}

func (h *%[1]sHandler) List(c *gin.Context) {
	p := pagination.Extract(c)
	list, meta, err := h.service.List%[1]s(c.Request.Context(), query.List%[1]sQuery{Pagination: p})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.Success(c, 200, "%[1]s list retrieved successfully", response.From%[1]sDTOList(list), meta)
}

func (h *%[1]sHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req request.Update%[1]sRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		sharedResponse.Error(c, appErrors.NewValidationError("invalid request payload", appErrors.FieldError{
			Field:   "body",
			Message: err.Error(),
		}))
		return
	}

	result, err := h.service.Update%[1]s(c.Request.Context(), command.Update%[1]sCommand{ID: id, Name: req.Name})
	if err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "%[1]s updated successfully", response.From%[1]sDTO(result))
}

func (h *%[1]sHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete%[1]s(c.Request.Context(), command.Delete%[1]sCommand{ID: id}); err != nil {
		sharedResponse.Error(c, err)
		return
	}

	sharedResponse.OK(c, "%[1]s deleted successfully", nil)
}
`, pascalName, featureName),

		// Delivery Routes
		filepath.Join(moduleDir, "delivery", "http", "routes.go"): fmt.Sprintf(`package http

import (
	authService "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/service"
	"%[3]s/application"
	"%[3]s/delivery/http/handler"
	rbacService "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/domain/service"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	apiGroup *gin.RouterGroup,
	tokenService authService.TokenService,
	authzService rbacService.AuthorizationService,
	service application.%[1]sService,
) {
	hdlr := handler.New%[1]sHandler(service)
	group := apiGroup.Group("/%[2]ss")
	{
		group.GET("", hdlr.List)
		group.GET("/:id", hdlr.GetByID)

		protected := group.Group("")
		protected.Use(middleware.AuthMiddleware(tokenService))
		{
			protected.POST("", middleware.RequirePermission(authzService, "%[2]s:create"), hdlr.Create)
			protected.PUT("/:id", middleware.RequirePermission(authzService, "%[2]s:update"), hdlr.Update)
			protected.DELETE("/:id", middleware.RequirePermission(authzService, "%[2]s:delete"), hdlr.Delete)
		}
	}
}
`, pascalName, featureName, fmt.Sprintf("github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/%s", featureName)),
	}

	for filePath, content := range files {
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
			return fmt.Errorf("failed to write scaffolded file %s: %w", filePath, err)
		}
	}

	return nil
}
