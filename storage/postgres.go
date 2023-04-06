package storage

import (
	"context"
	"database/sql"

	"github.com/Invan2/invan_catalog_service/config"
	"github.com/Invan2/invan_catalog_service/models"
	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/Invan2/invan_catalog_service/storage/postgres"
	"github.com/Invan2/invan_catalog_service/storage/repo"
	"github.com/jmoiron/sqlx"
)

type repos struct {
	productRepo         repo.ProductPgI
	categoryRepo        repo.CategoryPgI
	measurementUnitRepo repo.MeasurementUnitPgI
	companyRepo         repo.CompanyPgI
	userRepo            repo.UserPgI
	shopRepo            repo.ShopI
	label               repo.LabelI
	scalesTemplateRepo  repo.ScalesTemplateI
	supplierRepo        repo.SupplierI
	vatRepo             repo.VatI
}

type repoIs interface {
	Product() repo.ProductPgI
	MeasurementUnit() repo.MeasurementUnitPgI
	Category() repo.CategoryPgI
	Company() repo.CompanyPgI
	User() repo.UserPgI
	Shop() repo.ShopI
	Label() repo.LabelI
	ScalesTemplate() repo.ScalesTemplateI
	Supplier() repo.SupplierI
	Vat() repo.VatI
}

type storage struct {
	db  *sqlx.DB
	log logger.Logger
	cfg config.Config
	repos
}

type storageTr struct {
	tr *sqlx.Tx
	repos
}

type StorageTrI interface {
	Commit() error
	Rollback() error
	repoIs
}

type StoragePg interface {
	WithTransaction() (StorageTrI, error)
	repoIs
}

func NewStoragePg(log logger.Logger, db *sqlx.DB, cfg config.Config) StoragePg {

	return &storage{
		db:    db,
		log:   log,
		cfg:   cfg,
		repos: getRepos(log, db, cfg),
	}
}

func (s *storage) WithTransaction() (StorageTrI, error) {

	tr, err := s.db.BeginTxx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return nil, err
	}

	return &storageTr{
		tr:    tr,
		repos: getRepos(s.log, tr, s.cfg),
	}, nil
}

func getRepos(log logger.Logger, db models.DB, cfg config.Config) repos {
	return repos{
		productRepo:         postgres.NewProductRepo(log, db, cfg),
		measurementUnitRepo: postgres.NewMeasurementUnitRepo(log, db),
		categoryRepo:        postgres.NewCategoryRepo(log, db),
		companyRepo:         postgres.NewCompanyRepo(log, db),
		userRepo:            postgres.NewUserRepo(log, db),
		label:               postgres.NewLabelRepo(log, db),
		shopRepo:            postgres.NewShopRepo(log, db),
		scalesTemplateRepo:  postgres.NewScalesTemplateRepo(log, db, cfg),
		supplierRepo:        postgres.NewSupplierRepo(log, db, cfg),
		vatRepo:             postgres.NewVatRepo(log, db, cfg),
	}
}

func (s *storageTr) Commit() error {
	return s.tr.Commit()
}

func (s *storageTr) Rollback() error {
	return s.tr.Rollback()
}

func (r *repos) Product() repo.ProductPgI {
	return r.productRepo
}

func (r *repos) MeasurementUnit() repo.MeasurementUnitPgI {
	return r.measurementUnitRepo
}

func (r *repos) Category() repo.CategoryPgI {
	return r.categoryRepo
}

func (r *repos) Company() repo.CompanyPgI {
	return r.companyRepo
}

func (r *repos) User() repo.UserPgI {
	return r.userRepo
}

func (r *repos) Shop() repo.ShopI {
	return r.shopRepo
}

func (r *repos) Label() repo.LabelI {
	return r.label
}
func (r *repos) ScalesTemplate() repo.ScalesTemplateI {
	return r.scalesTemplateRepo
}

func (r *repos) Supplier() repo.SupplierI {
	return r.supplierRepo
}

func (r *repos) Vat() repo.VatI {
	return r.vatRepo
}
