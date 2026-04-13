package admin

// =============================================
// BRANCH DTOs
// =============================================

type CreateBranchRequest struct {
	Name    string `json:"name" validate:"required"`
	Code    string `json:"code" validate:"required"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	LogoURL string `json:"logo_url"`
}

type UpdateBranchRequest struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	LogoURL  string `json:"logo_url"`
	IsActive *bool  `json:"is_active"`
}

type BranchResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	LogoURL   string `json:"logo_url"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// =============================================
// USER DTOs
// =============================================

type CreateUserRequest struct {
	BranchID  string `json:"branch_id" validate:"required"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone"`
	Password  string `json:"password" validate:"required,min=6"`
	Role      string `json:"role" validate:"required,oneof=super_admin admin manager employee"`
}

type UpdateUserRequest struct {
	BranchID  string `json:"branch_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Role      string `json:"role" validate:"omitempty,oneof=super_admin admin manager employee"`
	Status    string `json:"status" validate:"omitempty,oneof=active inactive suspended pending"`
}

type ResetPasswordRequest struct {
	Password string `json:"password" validate:"required,min=6"`
}

type UserResponse struct {
	ID          string `json:"id"`
	BranchID    string `json:"branch_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Role        string `json:"role"`
	Status      string `json:"status"`
	AvatarURL   string `json:"avatar_url"`
	LastLoginAt string `json:"last_login_at"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// =============================================
// MENU DTOs
// =============================================

type CreateMenuRequest struct {
	ParentID  *string `json:"parent_id"`
	Label     string  `json:"label" validate:"required"`
	Path      string  `json:"path"`
	Resource  string  `json:"resource"`
	SortOrder int     `json:"sort_order"`
}

type UpdateMenuRequest struct {
	ParentID  *string `json:"parent_id"`
	Label     string  `json:"label"`
	Path      string  `json:"path"`
	Resource  string  `json:"resource"`
	SortOrder *int    `json:"sort_order"`
	IsActive  *bool   `json:"is_active"`
}

type MenuResponse struct {
	ID        string         `json:"id"`
	ParentID  *string        `json:"parent_id"`
	Label     string         `json:"label"`
	Path      string         `json:"path,omitempty"`
	Resource  string         `json:"resource,omitempty"`
	SortOrder int            `json:"sort_order"`
	IsActive  bool           `json:"is_active"`
	Children  []MenuResponse `json:"children,omitempty"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
}

// =============================================
// EMPLOYEE DTOs
// =============================================

type CreateEmployeeRequest struct {
	// User account fields
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone"`
	Password  string `json:"password" validate:"required,min=6"`
	// Employee profile fields
	BranchID       string   `json:"branch_id" validate:"required"`
	ManagerID      *string  `json:"manager_id"`
	EmployeeCode   string   `json:"employee_code" validate:"required"`
	Designation    string   `json:"designation"`
	EmploymentType string   `json:"employment_type" validate:"omitempty,oneof=full_time part_time contract"`
	HourlyRate     *float64 `json:"hourly_rate"`
	JoiningDate    string   `json:"joining_date" validate:"required"` // format: YYYY-MM-DD
}

type UpdateEmployeeRequest struct {
	// User fields
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Status    string `json:"status" validate:"omitempty,oneof=active inactive suspended pending"`
	// Employee fields
	ManagerID      *string  `json:"manager_id"`
	Designation    string   `json:"designation"`
	EmploymentType string   `json:"employment_type" validate:"omitempty,oneof=full_time part_time contract"`
	HourlyRate     *float64 `json:"hourly_rate"`
}

type EmployeeResponse struct {
	ID             string   `json:"id"`
	UserID         string   `json:"user_id"`
	BranchID       string   `json:"branch_id"`
	ManagerID      *string  `json:"manager_id,omitempty"`
	EmployeeCode   string   `json:"employee_code"`
	Designation    *string  `json:"designation,omitempty"`
	EmploymentType string   `json:"employment_type"`
	HourlyRate     *float64 `json:"hourly_rate,omitempty"`
	JoiningDate    string   `json:"joining_date"`
	FirstName      string   `json:"first_name"`
	LastName       string   `json:"last_name"`
	Email          string   `json:"email"`
	Phone          *string  `json:"phone,omitempty"`
	Status         string   `json:"status"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
}

// =============================================
// ROLE PERMISSION DTOs
// =============================================

type CreateRolePermissionRequest struct {
	Role      string `json:"role" validate:"required,oneof=super_admin admin manager employee"`
	Resource  string `json:"resource" validate:"required,oneof=employee attendance payroll report settings"`
	CanView   *bool  `json:"can_view"`
	CanCreate *bool  `json:"can_create"`
	CanEdit   *bool  `json:"can_edit"`
	CanDelete *bool  `json:"can_delete"`
}

type UpdateRolePermissionRequest struct {
	Role      string `json:"role" validate:"omitempty,oneof=super_admin admin manager employee"`
	Resource  string `json:"resource" validate:"omitempty,oneof=employee attendance payroll report settings"`
	CanView   *bool  `json:"can_view"`
	CanCreate *bool  `json:"can_create"`
	CanEdit   *bool  `json:"can_edit"`
	CanDelete *bool  `json:"can_delete"`
}
