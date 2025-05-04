package serviceAdmin

import (
	"complaint_server/internal/storage/pg"
	"context"
)

type AdminService struct {
	adminStorage *pg.Storage
}

func NewAdminService(adminStorage *pg.Storage) *AdminService {
	return &AdminService{adminStorage: adminStorage}
}

func (s *AdminService) IsAdmin(ctx context.Context, barcode int) (bool, error) {
	isAdmin, err := s.adminStorage.AdminExistsByBarcode(ctx, barcode)
	if err != nil {
		return false, err
	}

	return isAdmin, nil
}
