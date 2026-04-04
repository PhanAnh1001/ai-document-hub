package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/PhanAnh1001/ai-accounting/backend/internal/model"
)

type OrganizationRepository struct {
	pool *pgxpool.Pool
}

func NewOrganizationRepository(pool *pgxpool.Pool) *OrganizationRepository {
	return &OrganizationRepository{pool: pool}
}

func (r *OrganizationRepository) Create(ctx context.Context, org *model.Organization) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO organizations (id, name, tax_code, accounting_standard)
		 VALUES ($1, $2, $3, $4)`,
		org.ID, org.Name, org.TaxCode, org.AccountingStandard,
	)
	if err != nil {
		return fmt.Errorf("insert organization: %w", err)
	}
	return nil
}

func (r *OrganizationRepository) Update(ctx context.Context, org *model.Organization) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE organizations
		 SET name=$1, tax_code=$2, accounting_standard=$3,
		     ocr_provider=$4, misa_api_url=$5, misa_api_key=$6,
		     updated_at=now()
		 WHERE id=$7`,
		org.Name, org.TaxCode, org.AccountingStandard,
		org.OCRProvider, org.MISAAPIURL, org.MISAAPIKey,
		org.ID,
	)
	if err != nil {
		return fmt.Errorf("update organization: %w", err)
	}
	return nil
}

func (r *OrganizationRepository) GetByID(ctx context.Context, id string) (*model.Organization, error) {
	var o model.Organization
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, tax_code, accounting_standard, ocr_provider, misa_api_url, misa_api_key, created_at, updated_at
		 FROM organizations WHERE id = $1`, id,
	).Scan(&o.ID, &o.Name, &o.TaxCode, &o.AccountingStandard, &o.OCRProvider, &o.MISAAPIURL, &o.MISAAPIKey, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get organization: %w", err)
	}
	return &o, nil
}
