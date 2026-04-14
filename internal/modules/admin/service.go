package admin

import (
	"context"
	"errors"
	"time"

	"rmp-api/internal/models"
	"rmp-api/pkg/hash"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// =============================================
// BRANCH SERVICE
// =============================================

func (s *Service) GetAllBranches(ctx context.Context) ([]models.Branch, error) {
	return s.repo.FindAllBranches(ctx)
}

func (s *Service) GetBranchByID(ctx context.Context, id string) (*models.Branch, error) {
	return s.repo.FindBranchByID(ctx, id)
}

func (s *Service) CreateBranch(ctx context.Context, req CreateBranchRequest) (*models.Branch, error) {
	b := &models.Branch{
		Name: req.Name,
		Code: req.Code,
	}
	if req.Address != "" {
		b.Address = &req.Address
	}
	if req.Phone != "" {
		b.Phone = &req.Phone
	}
	if req.Email != "" {
		b.Email = &req.Email
	}
	if req.LogoURL != "" {
		b.LogoURL = &req.LogoURL
	}

	if err := s.repo.CreateBranch(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Service) UpdateBranch(ctx context.Context, id string, req UpdateBranchRequest) (*models.Branch, error) {
	existing, err := s.repo.FindBranchByID(ctx, id)
	if err != nil {
		return nil, errors.New("branch not found")
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Code != "" {
		existing.Code = req.Code
	}
	if req.Address != "" {
		existing.Address = &req.Address
	}
	if req.Phone != "" {
		existing.Phone = &req.Phone
	}
	if req.Email != "" {
		existing.Email = &req.Email
	}
	if req.LogoURL != "" {
		existing.LogoURL = &req.LogoURL
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateBranch(ctx, id, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) DeleteBranch(ctx context.Context, id string) error {
	return s.repo.DeleteBranch(ctx, id)
}

// =============================================
// USER SERVICE
// =============================================

func (s *Service) GetAllUsers(ctx context.Context) ([]models.User, error) {
	return s.repo.FindAllUsers(ctx)
}

func (s *Service) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return s.repo.FindUserByID(ctx, id)
}

func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) (*models.User, error) {
	passwordHash, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	u := &models.User{
		BranchID:     req.BranchID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         req.Role,
		Status:       "active",
	}
	if req.Phone != "" {
		u.Phone = &req.Phone
	}

	if err := s.repo.CreateUser(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (*models.User, error) {
	existing, err := s.repo.FindUserByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.BranchID != "" {
		existing.BranchID = req.BranchID
	}
	if req.FirstName != "" {
		existing.FirstName = req.FirstName
	}
	if req.LastName != "" {
		existing.LastName = req.LastName
	}
	if req.Email != "" {
		existing.Email = req.Email
	}
	if req.Phone != "" {
		existing.Phone = &req.Phone
	}
	if req.Role != "" {
		existing.Role = req.Role
	}
	if req.Status != "" {
		existing.Status = req.Status
	}

	if err := s.repo.UpdateUser(ctx, id, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) ResetUserPassword(ctx context.Context, id string, req ResetPasswordRequest) error {
	_, err := s.repo.FindUserByID(ctx, id)
	if err != nil {
		return errors.New("user not found")
	}

	passwordHash, err := hash.HashPassword(req.Password)
	if err != nil {
		return errors.New("failed to hash password")
	}

	return s.repo.UpdateUserPassword(ctx, id, passwordHash)
}

func (s *Service) DeleteUser(ctx context.Context, id string) error {
	return s.repo.DeleteUser(ctx, id)
}

// =============================================
// EMPLOYEE SERVICE
// =============================================

func (s *Service) GetAllEmployees(ctx context.Context) ([]models.Employee, error) {
	return s.repo.FindAllEmployees(ctx)
}

func (s *Service) GetEmployeeByID(ctx context.Context, id string) (*models.Employee, error) {
	return s.repo.FindEmployeeByID(ctx, id)
}

func (s *Service) CreateEmployee(ctx context.Context, req CreateEmployeeRequest) (*models.Employee, error) {
	passwordHash, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	u := &models.User{
		BranchID:     req.BranchID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PasswordHash: passwordHash,
	}
	if req.Phone != "" {
		u.Phone = &req.Phone
	}

	e := &models.Employee{
		BranchID:       req.BranchID,
		ManagerID:      req.ManagerID,
		EmployeeCode:   req.EmployeeCode,
		EmploymentType: "full_time",
		HourlyRate:     req.HourlyRate,
		Currency:       "USD",
	}
	if req.Designation != "" {
		e.Designation = &req.Designation
	}
	if req.EmploymentType != "" {
		e.EmploymentType = req.EmploymentType
	}
	if req.Currency != "" {
		e.Currency = req.Currency
	}

	joiningDate, err := parseDate(req.JoiningDate)
	if err != nil {
		return nil, errors.New("invalid joining_date format, use YYYY-MM-DD")
	}
	e.JoiningDate = joiningDate

	if err := s.repo.CreateEmployee(ctx, u, e); err != nil {
		return nil, err
	}

	e.FirstName = u.FirstName
	e.LastName = u.LastName
	e.Email = u.Email
	e.Phone = u.Phone
	e.Status = "active"
	return e, nil
}

func (s *Service) UpdateEmployee(ctx context.Context, id string, req UpdateEmployeeRequest) (*models.Employee, error) {
	existing, err := s.repo.FindEmployeeByID(ctx, id)
	if err != nil {
		return nil, errors.New("employee not found")
	}

	u := &models.User{
		FirstName: existing.FirstName,
		LastName:  existing.LastName,
		Phone:     existing.Phone,
		Status:    existing.Status,
	}
	if req.FirstName != "" {
		u.FirstName = req.FirstName
	}
	if req.LastName != "" {
		u.LastName = req.LastName
	}
	if req.Phone != "" {
		u.Phone = &req.Phone
	}
	if req.Status != "" {
		u.Status = req.Status
	}

	if req.ManagerID != nil {
		existing.ManagerID = req.ManagerID
	}
	if req.Designation != "" {
		existing.Designation = &req.Designation
	}
	if req.EmploymentType != "" {
		existing.EmploymentType = req.EmploymentType
	}
	if req.HourlyRate != nil {
		existing.HourlyRate = req.HourlyRate
	}
	if req.Currency != "" {
		existing.Currency = req.Currency
	}

	if err := s.repo.UpdateEmployee(ctx, id, u, existing); err != nil {
		return nil, err
	}

	existing.FirstName = u.FirstName
	existing.LastName = u.LastName
	existing.Phone = u.Phone
	existing.Status = u.Status
	return existing, nil
}

func (s *Service) DeleteEmployee(ctx context.Context, id string) error {
	return s.repo.DeleteEmployee(ctx, id)
}

// =============================================
// MENU SERVICE
// =============================================

func (s *Service) GetAllMenus(ctx context.Context) ([]models.Menu, error) {
	return s.repo.FindAllMenus(ctx)
}

func (s *Service) GetMenuByID(ctx context.Context, id string) (*models.Menu, error) {
	return s.repo.FindMenuByID(ctx, id)
}

func (s *Service) CreateMenu(ctx context.Context, req CreateMenuRequest) (*models.Menu, error) {
	m := &models.Menu{
		ParentID:  req.ParentID,
		Label:     req.Label,
		SortOrder: req.SortOrder,
	}
	if req.Path != "" {
		m.Path = &req.Path
	}
	if req.Resource != "" {
		m.Resource = &req.Resource
	}

	if err := s.repo.CreateMenu(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Service) UpdateMenu(ctx context.Context, id string, req UpdateMenuRequest) (*models.Menu, error) {
	existing, err := s.repo.FindMenuByID(ctx, id)
	if err != nil {
		return nil, errors.New("menu not found")
	}

	// ParentID can be explicitly set to null (move to top level)
	if req.ParentID != nil {
		existing.ParentID = req.ParentID
	}
	if req.Label != "" {
		existing.Label = req.Label
	}
	if req.Path != "" {
		existing.Path = &req.Path
	}
	if req.Resource != "" {
		existing.Resource = &req.Resource
	}
	if req.SortOrder != nil {
		existing.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateMenu(ctx, id, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) DeleteMenu(ctx context.Context, id string) error {
	return s.repo.DeleteMenu(ctx, id)
}

// =============================================
// ROLE PERMISSION SERVICE
// =============================================

func (s *Service) GetAllRolePermissions(ctx context.Context) ([]models.RolePermission, error) {
	return s.repo.FindAllRolePermissions(ctx)
}

func (s *Service) GetRolePermissionByID(ctx context.Context, id string) (*models.RolePermission, error) {
	return s.repo.FindRolePermissionByID(ctx, id)
}

func (s *Service) CreateRolePermission(ctx context.Context, req CreateRolePermissionRequest) (*models.RolePermission, error) {
	rp := &models.RolePermission{
		Role:     req.Role,
		Resource: req.Resource,
	}
	if req.CanView != nil {
		rp.CanView = *req.CanView
	}
	if req.CanCreate != nil {
		rp.CanCreate = *req.CanCreate
	}
	if req.CanEdit != nil {
		rp.CanEdit = *req.CanEdit
	}
	if req.CanDelete != nil {
		rp.CanDelete = *req.CanDelete
	}

	if err := s.repo.CreateRolePermission(ctx, rp); err != nil {
		return nil, err
	}
	return rp, nil
}

func (s *Service) UpdateRolePermission(ctx context.Context, id string, req UpdateRolePermissionRequest) (*models.RolePermission, error) {
	existing, err := s.repo.FindRolePermissionByID(ctx, id)
	if err != nil {
		return nil, errors.New("role permission not found")
	}

	if req.Role != "" {
		existing.Role = req.Role
	}
	if req.Resource != "" {
		existing.Resource = req.Resource
	}
	if req.CanView != nil {
		existing.CanView = *req.CanView
	}
	if req.CanCreate != nil {
		existing.CanCreate = *req.CanCreate
	}
	if req.CanEdit != nil {
		existing.CanEdit = *req.CanEdit
	}
	if req.CanDelete != nil {
		existing.CanDelete = *req.CanDelete
	}

	if err := s.repo.UpdateRolePermission(ctx, id, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) DeleteRolePermission(ctx context.Context, id string) error {
	return s.repo.DeleteRolePermission(ctx, id)
}

// GetAllMenusTree returns all menus as a nested tree (for admin view)
func (s *Service) GetAllMenusTree(ctx context.Context) ([]models.Menu, error) {
	flatMenus, err := s.repo.FindAllMenus(ctx)
	if err != nil {
		return nil, err
	}
	return BuildMenuTree(flatMenus), nil
}

// BuildMenuTree converts a flat list of menus into a nested tree structure
func BuildMenuTree(flatMenus []models.Menu) []models.Menu {
	menuMap := make(map[string]*models.Menu)
	var roots []models.Menu

	// Index all menus by ID
	for i := range flatMenus {
		flatMenus[i].Children = []models.Menu{}
		menuMap[flatMenus[i].ID] = &flatMenus[i]
	}

	// Build tree
	for i := range flatMenus {
		if flatMenus[i].ParentID != nil && *flatMenus[i].ParentID != "" {
			if parent, ok := menuMap[*flatMenus[i].ParentID]; ok {
				parent.Children = append(parent.Children, flatMenus[i])
			}
		} else {
			roots = append(roots, flatMenus[i])
		}
	}

	// Re-attach children to roots (since we copied by value)
	for i := range roots {
		if mapped, ok := menuMap[roots[i].ID]; ok {
			roots[i].Children = mapped.Children
		}
	}

	return roots
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}
