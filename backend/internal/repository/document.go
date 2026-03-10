package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/PhanAnh1001/ai-accounting/backend/internal/model"
)

type DocumentRepository struct {
	pool *pgxpool.Pool
}

func NewDocumentRepository(pool *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{pool: pool}
}

func (r *DocumentRepository) Create(ctx context.Context, doc *model.Document) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO documents (id, organization_id, file_url, file_name, document_type, status, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		doc.ID, doc.OrganizationID, doc.FileURL, doc.FileName, doc.DocumentType, doc.Status, doc.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("insert document: %w", err)
	}
	return nil
}

// UpdateStatus changes document processing status.
func (r *DocumentRepository) UpdateStatus(ctx context.Context, id string, status model.DocumentStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE documents SET status=$1, updated_at=now() WHERE id=$2`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("update document status: %w", err)
	}
	return nil
}

// UpdateOCRData stores OCR results and updates document status.
func (r *DocumentRepository) UpdateOCRData(ctx context.Context, id string, ocrData []byte, status model.DocumentStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE documents SET ocr_data=$1, status=$2, updated_at=now() WHERE id=$3`,
		ocrData, status, id,
	)
	if err != nil {
		return fmt.Errorf("update document ocr_data: %w", err)
	}
	return nil
}

// GetByID fetches a single document scoped to the organization.
func (r *DocumentRepository) GetByID(ctx context.Context, id, orgID string) (*model.Document, error) {
	var d model.Document
	err := r.pool.QueryRow(ctx,
		`SELECT id, organization_id, file_url, file_name, document_type, status, ocr_data, created_by, created_at, updated_at
		 FROM documents WHERE id=$1 AND organization_id=$2`,
		id, orgID,
	).Scan(&d.ID, &d.OrganizationID, &d.FileURL, &d.FileName,
		&d.DocumentType, &d.Status, &d.OCRData, &d.CreatedBy,
		&d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}
	return &d, nil
}

// CountByDay returns documents uploaded per day for the last N days, optionally filtered by document_type.
func (r *DocumentRepository) CountByDay(ctx context.Context, orgID string, days int, docType string) ([]model.DayCount, error) {
	where := "WHERE organization_id=$1 AND created_at >= NOW() - ($2 || ' days')::interval"
	args := []interface{}{orgID, days}
	if docType != "" {
		where += " AND document_type=$3"
		args = append(args, docType)
	}

	rows, err := r.pool.Query(ctx,
		fmt.Sprintf(`SELECT TO_CHAR(created_at::date, 'YYYY-MM-DD') AS day, COUNT(*) AS cnt
		 FROM documents %s
		 GROUP BY day
		 ORDER BY day ASC`, where),
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("count documents by day: %w", err)
	}
	defer rows.Close()

	var counts []model.DayCount
	for rows.Next() {
		var dc model.DayCount
		if err := rows.Scan(&dc.Date, &dc.Count); err != nil {
			return nil, fmt.Errorf("scan day count: %w", err)
		}
		counts = append(counts, dc)
	}
	return counts, nil
}

func (r *DocumentRepository) ListByOrg(ctx context.Context, orgID, status, q string, limit, offset int) ([]*model.Document, int, error) {
	where := "WHERE organization_id=$1"
	args := []interface{}{orgID}
	idx := 2

	if status != "" {
		where += fmt.Sprintf(" AND status=$%d", idx)
		args = append(args, status)
		idx++
	}
	if q != "" {
		where += fmt.Sprintf(" AND file_name ILIKE '%%' || $%d || '%%'", idx)
		args = append(args, q)
		idx++
	}

	var total int
	if err := r.pool.QueryRow(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM documents %s", where), args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count documents: %w", err)
	}

	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx,
		fmt.Sprintf(`SELECT id, organization_id, file_url, file_name, document_type, status, ocr_data, created_by, created_at, updated_at
		 FROM documents %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, idx, idx+1),
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("query documents: %w", err)
	}
	defer rows.Close()

	var docs []*model.Document
	for rows.Next() {
		var d model.Document
		if err := rows.Scan(
			&d.ID, &d.OrganizationID, &d.FileURL, &d.FileName,
			&d.DocumentType, &d.Status, &d.OCRData, &d.CreatedBy,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan document: %w", err)
		}
		docs = append(docs, &d)
	}
	return docs, total, nil
}
