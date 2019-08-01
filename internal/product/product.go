package product

import (
	"context"
	"time"

	"github.com/FrankSantoso/service/internal/platform/auth"
	"github.com/google/uuid"
	// "github.com/jmoiron/sqlx"
	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// Predefined errors identify expected failure conditions.
var (
	// ErrNotFound is used when a specific Product is requested but does not exist.
	ErrNotFound = errors.New("Product not found")

	// ErrInvalidID is used when an invalid UUID is provided.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrForbidden occurs when a user tries to do something that is forbidden to
	// them according to our access control policies.
	ErrForbidden = errors.New("Attempted action is not allowed")
)

// List gets all Products from the database.
func List(ctx context.Context, db *pg.DB) ([]Product, error) {
	ctx, span := trace.StartSpan(ctx, "internal.product.List")
	defer span.End()

	products := []Product{}

	err := db.ModelContext(ctx, &products).
		ColumnExpr("p.*").
		ColumnExpr("COALESCE(SUM(s.quantity),0) AS sold").
		ColumnExpr("COALESCE(SUM(s.paid), 0) AS revenue").
		Join("LEFT JOIN sales AS s ON s.product_id = p.product_id").
		Group("p.product_id").
		Select()

	if err != nil {
		return nil, errors.Wrap(err, "selecting products")
	}

	return products, nil
}

// Create adds a Product to the database. It returns the created Product with
// fields like ID and DateCreated populated..
func Create(ctx context.Context, db *pg.DB, user auth.Claims, np NewProduct, now time.Time) (*Product, error) {
	ctx, span := trace.StartSpan(ctx, "internal.product.Create")
	defer span.End()

	p := Product{
		ID:          uuid.New().String(),
		Name:        np.Name,
		Cost:        np.Cost,
		Quantity:    np.Quantity,
		Sold:        0,
		UserID:      user.Subject,
		DateCreated: now.UTC(),
		DateUpdated: now.UTC(),
	}

	err := db.RunInTransaction(func(tx *pg.Tx) error {
		if _, err := tx.Model(&p).Insert(); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "inserting product")
	}

	return &p, nil
}

// Retrieve finds the product identified by a given ID.
func Retrieve(ctx context.Context, db *pg.DB, id string) (*Product, error) {
	ctx, span := trace.StartSpan(ctx, "internal.product.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	var p Product

	err := db.ModelContext(ctx, &p).
		ColumnExpr("p.*").
		ColumnExpr("COALESCE(SUM(s.quantity),0) AS sold").
		ColumnExpr("COALESCE(SUM(s.paid), 0) AS revenue").
		Join("LEFT JOIN sales AS s ON s.product_id = p.product_id").
		Group("p.product_id").
		Where("p.product_id = ?", id).
		First()

	if err != nil {
		if err == pg.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, "retrieving single product")
	}

	return &p, nil
}

// Update modifies data about a Product. It will error if the specified ID is
// invalid or does not reference an existing Product.
func Update(ctx context.Context, db *pg.DB, user auth.Claims, id string, update UpdateProduct, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.product.Update")
	defer span.End()

	p, err := Retrieve(ctx, db, id)
	if err != nil {
		return err
	}

	// If you do not have the admin role ...
	// and you are not the owner of this product ...
	// then get outta here!
	if !user.HasRole(auth.RoleAdmin) && p.UserID != user.Subject {
		return ErrForbidden
	}

	if update.Name != nil {
		p.Name = *update.Name
	}
	if update.Cost != nil {
		p.Cost = *update.Cost
	}
	if update.Quantity != nil {
		p.Quantity = *update.Quantity
	}
	p.DateUpdated = now

	err = db.RunInTransaction(func(tx *pg.Tx) error {
		if _, err := tx.ModelContext(ctx, p).Where("product_id = ?", id).
			UpdateNotZero(); err != nil {
			return err
		}
		return nil
	})

	// _, err = db.ModelContext(ctx, p).Where("product_id = ?", id).UpdateNotZero()

	if err != nil {
		return errors.Wrap(err, "updating product")
	}

	return nil
}

// Delete removes the product identified by a given ID.
func Delete(ctx context.Context, db *pg.DB, id string) error {
	ctx, span := trace.StartSpan(ctx, "internal.product.Delete")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}

	var p = new(Product)

	err := db.RunInTransaction(func(tx *pg.Tx) error {
		if _, err := tx.ModelContext(ctx, p).Where("product_id = ?", id).Delete(); err != nil {
			if err == pg.ErrNoRows {
				return ErrNotFound
			}
			return err
		}
		return nil
	})

	if err != nil {
		return errors.Wrapf(err, "deleting product %s", id)
	}

	return nil
}
