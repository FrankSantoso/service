package product

import (
	"github.com/FrankSantoso/service/internal/user"
	"time"
)

// Product is an item we sell.
type Product struct {
	tableName   struct{}   `sql:"alias:p"`
	ID          string     `sql:"product_id,pk" json:"id"`          // Unique identifier.
	Name        string     `sql:"name" json:"name"`                 // Display name of the product.
	Cost        int        `sql:"cost" json:"cost"`                 // Price for one item in cents.
	Quantity    int        `sql:"quantity" json:"quantity"`         // Original number of items available.
	Sold        int        `sql:"-" json:"sold"`                    // Aggregate field showing number of items sold.
	Revenue     int        `sql:"-" json:"revenue"`                 // Aggregate field showing total cost of sold items.
	UserID      string     `sql:"user_id" json:"user_id"`           // ID of the user who created the product.
	User        *user.User `sql:"-" pg:"fk:user_id" json:"user"`    // Specifies user_id as foreign key
	DateCreated time.Time  `sql:"date_created" json:"date_created"` // When the product was added.
	DateUpdated time.Time  `sql:"date_updated" json:"date_updated"` // When the product record was last modified.
}

// NewProduct is what we require from clients when adding a Product.
type NewProduct struct {
	Name     string `json:"name" validate:"required"`
	Cost     int    `json:"cost" validate:"required,gte=0"`
	Quantity int    `json:"quantity" validate:"gte=1"`
}

// UpdateProduct defines what information may be provided to modify an
// existing Product. All fields are optional so clients can send just the
// fields they want changed. It uses pointer fields so we can differentiate
// between a field that was not provided and a field that was provided as
// explicitly blank. Normally we do not want to use pointers to basic types but
// we make exceptions around marshalling/unmarshalling.
type UpdateProduct struct {
	Name     *string `json:"name"`
	Cost     *int    `json:"cost" validate:"omitempty,gte=0"`
	Quantity *int    `json:"quantity" validate:"omitempty,gte=1"`
}

// Sale represents one item of a transaction where some amount of a product was
// sold. Quantity is the number of units sold and Paid is the total price paid.
// Note that due to haggling the Paid value might not equal Quantity sold *
// Product cost.
type Sale struct {
	tableName   struct{}  `sql:"sales,alias:s"`
	ID          string    `sql:"sale_id,pk" json:"id"`
	ProductID   string    `sql:"product_id" json:"product_id"`
	Quantity    int       `sql:"quantity" json:"quantity"`
	Paid        int       `sql:"paid" json:"paid"`
	DateCreated time.Time `sql:"date_created" json:"date_created"`
}

// NewSale is what we require from clients for recording new transactions.
type NewSale struct {
	Quantity int `json:"quantity" validate:"gte=0"`
	Paid     int `json:"paid" validate:"gte=0"`
}
