package dbscope

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GetEffectiveBranchIDs returns the branch IDs the caller is authorised to operate on.
//
//   - nil slice  → super_admin; no branch filter should be applied
//   - non-nil    → restrict queries to these branch IDs
//
// For regional_manager the list is built from regional_manager_branches plus the
// home branch so the caller always has access to at least their own branch.
func GetEffectiveBranchIDs(ctx context.Context, db *pgxpool.Pool, role, userID, homeBranchID string) ([]string, error) {
	if role == "super_admin" {
		return nil, nil
	}
	if role != "regional_manager" {
		return []string{homeBranchID}, nil
	}

	rows, err := db.Query(ctx,
		`SELECT branch_id FROM regional_manager_branches WHERE regional_manager_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seen := map[string]bool{homeBranchID: true}
	ids := []string{homeBranchID}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// ContainsBranch reports whether branchID is in the allowed set.
// A nil slice (super_admin) grants access to every branch.
func ContainsBranch(branchIDs []string, branchID string) bool {
	if branchIDs == nil {
		return true
	}
	for _, id := range branchIDs {
		if id == branchID {
			return true
		}
	}
	return false
}
