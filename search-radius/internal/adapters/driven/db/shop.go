package db

import (
	"context"

	"search-radius/internal/adapters/driven/db/ent/generate"
	"search-radius/internal/adapters/driven/db/ent/generate/shop"
	"search-radius/internal/adapters/driven/db/mapper"
	"search-radius/internal/core/entity"
	"search-radius/internal/ports"
	"search-radius/pkg/database/ent"
	d "search-radius/pkg/dto"

	"entgo.io/ent/dialect/sql"
)

type ShopRepository struct {
	repo   *ent.BaseRepository[generate.Shop, *generate.Shop, int]
	client *generate.ShopClient
}

func NewShop(client interface{}) ports.ShopRepository {
	entClient := client.(*generate.Client)
	return &ShopRepository{
		repo:   ent.NewBaseRepository[generate.Shop, *generate.Shop, int](client),
		client: entClient.Shop,
	}
}

func (r *ShopRepository) Find(ctx context.Context, opts *d.QueryOptions) (*d.Paginated[*entity.Shop], error) {
	result, err := r.repo.Find(ctx, opts)
	if err != nil {
		return nil, err
	}

	entities := make([]*entity.Shop, len(*result.Records))
	for i, record := range *result.Records {
		entities[i] = mapper.ToShopEntity(record)
	}

	return &d.Paginated[*entity.Shop]{
		Records:    &entities,
		Pagination: result.Pagination,
	}, nil
}

func (r *ShopRepository) Get(ctx context.Context, id int) (*entity.Shop, error) {
	record, err := r.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapper.ToShopEntity(record), nil
}

func (r *ShopRepository) GetByIDs(ctx context.Context, ids []int) ([]*entity.Shop, error) {
	records, err := r.client.Query().Where(shop.IDIn(ids...)).All(ctx)
	if err != nil {
		return nil, err
	}

	entities := make([]*entity.Shop, len(records))
	for i, record := range records {
		entities[i] = mapper.ToShopEntity(record)
	}
	return entities, nil
}

func (r *ShopRepository) Create(ctx context.Context, e *entity.Shop) error {
	model := mapper.ToShopModel(e)
	if err := r.repo.Create(ctx, model); err != nil {
		return err
	}

	if created := mapper.ToShopEntity(model); created != nil {
		*e = *created
	}
	return nil
}

func (r *ShopRepository) CreateBatch(ctx context.Context, shops []*entity.Shop) error {
	builders := make([]*generate.ShopCreate, len(shops))
	for i, s := range shops {
		builders[i] = r.client.Create().
			SetName(s.Name).
			SetLat(s.Lat).
			SetLng(s.Lng)
	}
	_, err := r.client.CreateBulk(builders...).Save(ctx)
	return err
}

func (r *ShopRepository) Update(ctx context.Context, id int, e *entity.Shop) error {
	model := mapper.ToShopModel(e)
	model.ID = id
	if err := r.repo.Update(ctx, model); err != nil {
		return err
	}
	e.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *ShopRepository) Delete(ctx context.Context, id int) error {
	return r.repo.Delete(ctx, id)
}

func (r *ShopRepository) Exists(ctx context.Context, id int) (bool, error) {
	return r.repo.Exists(ctx, id)
}

func (r *ShopRepository) SearchRadius(ctx context.Context, lat, lng, radius float64) ([]*entity.Shop, error) {
	const metersPerKm = 1000

	// ST_Distance_Sphere(POINT(g.lng, g.lat), POINT(lng, lat)) <= radius_in_meters
	shops, err := r.client.Query().
		Where(func(s *sql.Selector) {
			s.Where(sql.ExprP(
				"ST_Distance_Sphere(POINT("+s.C(shop.FieldLng)+", "+s.C(shop.FieldLat)+"), POINT(?, ?)) <= ?",
				lng, lat, radius*metersPerKm,
			))
		}).
		All(ctx)

	if err != nil {
		return nil, err
	}

	entities := make([]*entity.Shop, len(shops))
	for i, record := range shops {
		entities[i] = mapper.ToShopEntity(record)
	}

	return entities, nil
}
