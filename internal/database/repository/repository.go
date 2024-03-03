package repository

type Product struct {
	Id       Id `bson:"_id"`
	Category string
	Cost     float64
	Quantity int
}

type CartElement struct {
	Id       Id      `bson:"_id"`
	Category string  `bson:"category"`
	Cost     float64 `bson:"cost"`
}

type Id struct {
	Name         string `bson:"name"`
	Manufacturer string `bson:"manufacturer"`
}

type ProductFilter struct {
	Category    string
	Keyword     string
	FloorCost   float64
	CeilingCost float64
}

type ProductRepository interface {
	GetProducts(filter ProductFilter) ([]Product, error)
	AddProductToCart(clientName string, clientPhone string, product Product) ([]CartElement, error)
}
