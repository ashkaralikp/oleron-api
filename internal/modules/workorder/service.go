package workorder

import (
	"context"
	"errors"
	"strings"
	"time"

	"rmp-api/internal/dbscope"
	"rmp-api/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) UpsertAsset(ctx context.Context, regionalManagerID, signerName string, signerTitle *string, signatureURL, sealURL string, uploadedBy *string) (*models.RegionalManagerWorkOrderAsset, error) {
	signerName = strings.TrimSpace(signerName)
	if signerName == "" {
		return nil, errors.New("signer_name is required")
	}
	signerTitle = trimOptional(signerTitle)
	return s.repo.UpsertAsset(ctx, regionalManagerID, signerName, signerTitle, signatureURL, sealURL, uploadedBy)
}

func (s *Service) GetAsset(ctx context.Context, regionalManagerID string) (*models.RegionalManagerWorkOrderAsset, error) {
	return s.repo.FindAssetByManagerID(ctx, regionalManagerID)
}

func (s *Service) Create(ctx context.Context, branchIDs []string, homeBranchID, userID, role string, req CreateWorkOrderRequest) (*models.WorkOrder, error) {
	targetBranch := homeBranchID
	if req.BranchID != "" {
		targetBranch = req.BranchID
	}
	if targetBranch == "" {
		return nil, errors.New("branch_id is required")
	}
	if !dbscope.ContainsBranch(branchIDs, targetBranch) {
		return nil, errors.New("forbidden")
	}
	if err := normalizeCreateRequest(&req); err != nil {
		return nil, err
	}

	var asset *models.RegionalManagerWorkOrderAsset
	if role == "regional_manager" {
		var err error
		asset, err = s.repo.FindAssetByManagerID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if asset == nil {
			return nil, errors.New("work order signature and seal must be uploaded first")
		}
	}

	return s.repo.CreateWorkOrder(ctx, targetBranch, userID, asset, req)
}

func (s *Service) GetAll(ctx context.Context, branchIDs []string) ([]models.WorkOrder, error) {
	return s.repo.FindAllWorkOrders(ctx, branchIDs)
}

func (s *Service) GetByID(ctx context.Context, id string, branchIDs []string) (*models.WorkOrder, error) {
	wo, err := s.repo.FindWorkOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if wo == nil {
		return nil, errors.New("work order not found")
	}
	if !dbscope.ContainsBranch(branchIDs, wo.BranchID) {
		return nil, errors.New("forbidden")
	}
	return wo, nil
}

func (s *Service) Update(ctx context.Context, id string, branchIDs []string, req UpdateWorkOrderRequest) (*models.WorkOrder, error) {
	if err := normalizeUpdateRequest(&req); err != nil {
		return nil, err
	}
	if _, err := s.GetByID(ctx, id, branchIDs); err != nil {
		return nil, err
	}
	wo, err := s.repo.UpdateWorkOrder(ctx, id, req)
	if err != nil {
		return nil, err
	}
	if wo == nil {
		return nil, errors.New("work order not found")
	}
	return wo, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id string, branchIDs []string, req UpdateWorkOrderStatusRequest) (*models.WorkOrder, error) {
	if _, err := s.GetByID(ctx, id, branchIDs); err != nil {
		return nil, err
	}
	wo, err := s.repo.UpdateStatus(ctx, id, req.Status, trimOptional(req.PDFURL))
	if err != nil {
		return nil, err
	}
	if wo == nil {
		return nil, errors.New("work order not found")
	}
	return wo, nil
}

func (s *Service) Delete(ctx context.Context, id string, branchIDs []string) error {
	wo, err := s.GetByID(ctx, id, branchIDs)
	if err != nil {
		return err
	}
	if wo.Status != "draft" {
		return errors.New("only draft work orders can be deleted")
	}
	deleted, err := s.repo.DeleteWorkOrder(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return errors.New("work order not found")
	}
	return nil
}

func normalizeCreateRequest(req *CreateWorkOrderRequest) error {
	req.CompanyName = strings.TrimSpace(req.CompanyName)
	req.BillToName = strings.TrimSpace(req.BillToName)
	req.JobDetails = strings.TrimSpace(req.JobDetails)
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	req.CompanyAddress = trimOptional(req.CompanyAddress)
	req.CompanyPhone = trimOptional(req.CompanyPhone)
	req.CompanyFax = trimOptional(req.CompanyFax)
	req.CompanyEmail = trimOptionalLower(req.CompanyEmail)
	req.CompanyWebsite = trimOptional(req.CompanyWebsite)
	req.CompanyLogoURL = trimOptional(req.CompanyLogoURL)
	req.BillToAddress = trimOptional(req.BillToAddress)
	req.BillToEmail = trimOptionalLower(req.BillToEmail)

	if req.CompanyName == "" {
		req.CompanyName = "Oleron.Inc"
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}
	if _, err := time.Parse("2006-01-02", req.WorkOrderDate); err != nil {
		return errors.New("invalid work_order_date: use YYYY-MM-DD")
	}
	for i := range req.Items {
		req.Items[i].Description = strings.TrimSpace(req.Items[i].Description)
		if req.Items[i].Description == "" {
			return errors.New("item description is required")
		}
	}
	if hasDuplicateLineNo(req.Items) {
		return errors.New("duplicate item line_no")
	}
	return nil
}

func normalizeUpdateRequest(req *UpdateWorkOrderRequest) error {
	req.CompanyAddress = trimOptional(req.CompanyAddress)
	req.CompanyPhone = trimOptional(req.CompanyPhone)
	req.CompanyFax = trimOptional(req.CompanyFax)
	req.CompanyEmail = trimOptionalLower(req.CompanyEmail)
	req.CompanyWebsite = trimOptional(req.CompanyWebsite)
	req.CompanyLogoURL = trimOptional(req.CompanyLogoURL)
	req.BillToAddress = trimOptional(req.BillToAddress)
	req.BillToEmail = trimOptionalLower(req.BillToEmail)
	req.PDFURL = trimOptional(req.PDFURL)

	if req.WorkOrderDate != nil {
		*req.WorkOrderDate = strings.TrimSpace(*req.WorkOrderDate)
		if _, err := time.Parse("2006-01-02", *req.WorkOrderDate); err != nil {
			return errors.New("invalid work_order_date: use YYYY-MM-DD")
		}
	}
	if req.CompanyName != nil {
		*req.CompanyName = strings.TrimSpace(*req.CompanyName)
		if *req.CompanyName == "" {
			return errors.New("company_name cannot be empty")
		}
	}
	if req.BillToName != nil {
		*req.BillToName = strings.TrimSpace(*req.BillToName)
		if *req.BillToName == "" {
			return errors.New("bill_to_name cannot be empty")
		}
	}
	if req.JobDetails != nil {
		*req.JobDetails = strings.TrimSpace(*req.JobDetails)
		if *req.JobDetails == "" {
			return errors.New("job_details cannot be empty")
		}
	}
	if req.Currency != nil {
		*req.Currency = strings.ToUpper(strings.TrimSpace(*req.Currency))
	}
	for i := range req.Items {
		req.Items[i].Description = strings.TrimSpace(req.Items[i].Description)
		if req.Items[i].Description == "" {
			return errors.New("item description is required")
		}
	}
	if hasDuplicateLineNo(req.Items) {
		return errors.New("duplicate item line_no")
	}
	return nil
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func trimOptionalLower(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.ToLower(strings.TrimSpace(*value))
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func hasDuplicateLineNo(items []WorkOrderItemRequest) bool {
	seen := map[int]bool{}
	for i, item := range items {
		lineNo := item.LineNo
		if lineNo == 0 {
			lineNo = i + 1
		}
		if seen[lineNo] {
			return true
		}
		seen[lineNo] = true
	}
	return false
}
