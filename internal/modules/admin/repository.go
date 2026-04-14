package admin

import (
	"context"

	"rmp-api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// =============================================
// BRANCH REPOSITORY
// =============================================

func (r *Repository) FindAllBranches(ctx context.Context) ([]models.Branch, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, code, address, phone, email, logo_url, is_active, created_at, updated_at
		 FROM branches ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []models.Branch
	for rows.Next() {
		var b models.Branch
		err := rows.Scan(
			&b.ID, &b.Name, &b.Code, &b.Address, &b.Phone,
			&b.Email, &b.LogoURL, &b.IsActive, &b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		branches = append(branches, b)
	}
	return branches, nil
}

func (r *Repository) FindBranchByID(ctx context.Context, id string) (*models.Branch, error) {
	var b models.Branch
	err := r.db.QueryRow(ctx,
		`SELECT id, name, code, address, phone, email, logo_url, is_active, created_at, updated_at
		 FROM branches WHERE id = $1`, id,
	).Scan(
		&b.ID, &b.Name, &b.Code, &b.Address, &b.Phone,
		&b.Email, &b.LogoURL, &b.IsActive, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *Repository) CreateBranch(ctx context.Context, b *models.Branch) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO branches (name, code, address, phone, email, logo_url)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, is_active, created_at, updated_at`,
		b.Name, b.Code, b.Address, b.Phone, b.Email, b.LogoURL,
	).Scan(&b.ID, &b.IsActive, &b.CreatedAt, &b.UpdatedAt)
}

func (r *Repository) UpdateBranch(ctx context.Context, id string, b *models.Branch) error {
	_, err := r.db.Exec(ctx,
		`UPDATE branches
		 SET name = $2, code = $3, address = $4, phone = $5,
		     email = $6, logo_url = $7, is_active = $8
		 WHERE id = $1`,
		id, b.Name, b.Code, b.Address, b.Phone, b.Email, b.LogoURL, b.IsActive,
	)
	return err
}

func (r *Repository) DeleteBranch(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM branches WHERE id = $1`, id)
	return err
}

// =============================================
// USER REPOSITORY
// =============================================

func (r *Repository) FindAllUsers(ctx context.Context) ([]models.User, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, branch_id, first_name, last_name, email, phone,
		        password_hash, role, status, avatar_url, last_login_at,
		        created_at, updated_at
		 FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(
			&u.ID, &u.BranchID, &u.FirstName, &u.LastName,
			&u.Email, &u.Phone, &u.PasswordHash, &u.Role,
			&u.Status, &u.AvatarURL, &u.LastLoginAt,
			&u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *Repository) FindUserByID(ctx context.Context, id string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow(ctx,
		`SELECT id, branch_id, first_name, last_name, email, phone,
		        password_hash, role, status, avatar_url, last_login_at,
		        created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(
		&u.ID, &u.BranchID, &u.FirstName, &u.LastName,
		&u.Email, &u.Phone, &u.PasswordHash, &u.Role,
		&u.Status, &u.AvatarURL, &u.LastLoginAt,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) CreateUser(ctx context.Context, u *models.User) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO users (branch_id, first_name, last_name, email, phone, password_hash, role, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, created_at, updated_at`,
		u.BranchID, u.FirstName, u.LastName, u.Email,
		u.Phone, u.PasswordHash, u.Role, u.Status,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *Repository) UpdateUser(ctx context.Context, id string, u *models.User) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users
		 SET branch_id = $2, first_name = $3, last_name = $4, email = $5,
		     phone = $6, role = $7, status = $8
		 WHERE id = $1`,
		id, u.BranchID, u.FirstName, u.LastName, u.Email,
		u.Phone, u.Role, u.Status,
	)
	return err
}

func (r *Repository) UpdateUserPassword(ctx context.Context, id string, passwordHash string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET password_hash = $2 WHERE id = $1`,
		id, passwordHash,
	)
	return err
}

func (r *Repository) DeleteUser(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

// =============================================
// EMPLOYEE REPOSITORY
// =============================================

func (r *Repository) FindAllEmployees(ctx context.Context, branchID string) ([]models.Employee, error) {
	query := `SELECT e.id, e.user_id, e.branch_id, e.manager_id, e.employee_code,
		        e.designation, e.employment_type, e.hourly_rate, e.currency, e.joining_date,
		        e.created_at, e.updated_at,
		        u.first_name, u.last_name, u.email, u.phone, u.status, u.avatar_url
		 FROM employees e
		 JOIN users u ON u.id = e.user_id`

	var args []any
	if branchID != "" {
		query += ` WHERE e.branch_id = $1`
		args = append(args, branchID)
	}
	query += ` ORDER BY e.created_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []models.Employee
	for rows.Next() {
		var e models.Employee
		err := rows.Scan(
			&e.ID, &e.UserID, &e.BranchID, &e.ManagerID, &e.EmployeeCode,
			&e.Designation, &e.EmploymentType, &e.HourlyRate, &e.Currency, &e.JoiningDate,
			&e.CreatedAt, &e.UpdatedAt,
			&e.FirstName, &e.LastName, &e.Email, &e.Phone, &e.Status, &e.AvatarURL,
		)
		if err != nil {
			return nil, err
		}
		employees = append(employees, e)
	}
	return employees, nil
}

func (r *Repository) FindEmployeeByID(ctx context.Context, id string) (*models.Employee, error) {
	var e models.Employee
	err := r.db.QueryRow(ctx,
		`SELECT e.id, e.user_id, e.branch_id, e.manager_id, e.employee_code,
		        e.designation, e.employment_type, e.hourly_rate, e.currency, e.joining_date,
		        e.created_at, e.updated_at,
		        u.first_name, u.last_name, u.email, u.phone, u.status, u.avatar_url
		 FROM employees e
		 JOIN users u ON u.id = e.user_id
		 WHERE e.id = $1`, id,
	).Scan(
		&e.ID, &e.UserID, &e.BranchID, &e.ManagerID, &e.EmployeeCode,
		&e.Designation, &e.EmploymentType, &e.HourlyRate, &e.Currency, &e.JoiningDate,
		&e.CreatedAt, &e.UpdatedAt,
		&e.FirstName, &e.LastName, &e.Email, &e.Phone, &e.Status, &e.AvatarURL,
	)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *Repository) CreateEmployee(ctx context.Context, u *models.User, e *models.Employee) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Create user account with employee role
	err = tx.QueryRow(ctx,
		`INSERT INTO users (branch_id, first_name, last_name, email, phone, password_hash, role, status)
		 VALUES ($1, $2, $3, $4, $5, $6, 'employee', 'active')
		 RETURNING id, created_at, updated_at`,
		u.BranchID, u.FirstName, u.LastName, u.Email,
		u.Phone, u.PasswordHash,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return err
	}

	// Create employee profile
	err = tx.QueryRow(ctx,
		`INSERT INTO employees (user_id, branch_id, manager_id, employee_code, designation, employment_type, hourly_rate, currency, joining_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, created_at, updated_at`,
		u.ID, e.BranchID, e.ManagerID, e.EmployeeCode,
		e.Designation, e.EmploymentType, e.HourlyRate, e.Currency, e.JoiningDate,
	).Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return err
	}

	e.UserID = u.ID
	return tx.Commit(ctx)
}

func (r *Repository) UpdateEmployee(ctx context.Context, id string, u *models.User, e *models.Employee) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`UPDATE users SET first_name = $2, last_name = $3, phone = $4, status = $5 WHERE id = $1`,
		e.UserID, u.FirstName, u.LastName, u.Phone, u.Status,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`UPDATE employees
		 SET manager_id = $2, designation = $3, employment_type = $4, hourly_rate = $5, currency = $6
		 WHERE id = $1`,
		id, e.ManagerID, e.Designation, e.EmploymentType, e.HourlyRate, e.Currency,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) DeleteEmployee(ctx context.Context, id string) error {
	// Deleting the employee row cascades via user_id → users ON DELETE CASCADE
	_, err := r.db.Exec(ctx,
		`DELETE FROM users WHERE id = (SELECT user_id FROM employees WHERE id = $1)`, id)
	return err
}

// =============================================
// MENU REPOSITORY
// =============================================

func (r *Repository) FindAllMenus(ctx context.Context) ([]models.Menu, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, parent_id, label, path, resource, sort_order, is_active, created_at, updated_at
		 FROM menus ORDER BY sort_order ASC, created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []models.Menu
	for rows.Next() {
		var m models.Menu
		err := rows.Scan(
			&m.ID, &m.ParentID, &m.Label, &m.Path,
			&m.Resource, &m.SortOrder, &m.IsActive, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		menus = append(menus, m)
	}
	return menus, nil
}

func (r *Repository) FindMenuByID(ctx context.Context, id string) (*models.Menu, error) {
	var m models.Menu
	err := r.db.QueryRow(ctx,
		`SELECT id, parent_id, label, path, resource, sort_order, is_active, created_at, updated_at
		 FROM menus WHERE id = $1`, id,
	).Scan(
		&m.ID, &m.ParentID, &m.Label, &m.Path,
		&m.Resource, &m.SortOrder, &m.IsActive, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) CreateMenu(ctx context.Context, m *models.Menu) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO menus (parent_id, label, path, resource, sort_order)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, is_active, created_at, updated_at`,
		m.ParentID, m.Label, m.Path, m.Resource, m.SortOrder,
	).Scan(&m.ID, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
}

func (r *Repository) UpdateMenu(ctx context.Context, id string, m *models.Menu) error {
	_, err := r.db.Exec(ctx,
		`UPDATE menus
		 SET parent_id = $2, label = $3, path = $4,
		     resource = $5, sort_order = $6, is_active = $7
		 WHERE id = $1`,
		id, m.ParentID, m.Label, m.Path, m.Resource, m.SortOrder, m.IsActive,
	)
	return err
}

func (r *Repository) DeleteMenu(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM menus WHERE id = $1`, id)
	return err
}

// =============================================
// ROLE PERMISSION REPOSITORY
// =============================================

func (r *Repository) FindAllRolePermissions(ctx context.Context) ([]models.RolePermission, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, role, resource, can_view, can_create, can_edit, can_delete, created_at
		 FROM role_permissions
		 ORDER BY role ASC, resource ASC, created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []models.RolePermission
	for rows.Next() {
		var rp models.RolePermission
		err := rows.Scan(
			&rp.ID, &rp.Role, &rp.Resource,
			&rp.CanView, &rp.CanCreate, &rp.CanEdit, &rp.CanDelete, &rp.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, rp)
	}
	return permissions, nil
}

func (r *Repository) FindRolePermissionByID(ctx context.Context, id string) (*models.RolePermission, error) {
	var rp models.RolePermission
	err := r.db.QueryRow(ctx,
		`SELECT id, role, resource, can_view, can_create, can_edit, can_delete, created_at
		 FROM role_permissions
		 WHERE id = $1`, id,
	).Scan(
		&rp.ID, &rp.Role, &rp.Resource,
		&rp.CanView, &rp.CanCreate, &rp.CanEdit, &rp.CanDelete, &rp.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rp, nil
}

func (r *Repository) CreateRolePermission(ctx context.Context, rp *models.RolePermission) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO role_permissions (role, resource, can_view, can_create, can_edit, can_delete)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		rp.Role, rp.Resource, rp.CanView, rp.CanCreate, rp.CanEdit, rp.CanDelete,
	).Scan(&rp.ID, &rp.CreatedAt)
}

func (r *Repository) UpdateRolePermission(ctx context.Context, id string, rp *models.RolePermission) error {
	_, err := r.db.Exec(ctx,
		`UPDATE role_permissions
		 SET role = $2, resource = $3, can_view = $4, can_create = $5, can_edit = $6, can_delete = $7
		 WHERE id = $1`,
		id, rp.Role, rp.Resource, rp.CanView, rp.CanCreate, rp.CanEdit, rp.CanDelete,
	)
	return err
}

func (r *Repository) DeleteRolePermission(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM role_permissions WHERE id = $1`, id)
	return err
}
