package store_domain

type Product struct {
	ID            string
	Name          string
	ActualPrice   float32
	DiscountPrice float32
}

func NewProduct(id, name string, actualPrice, discount float32) Product {
	return Product{
		ID:            id,
		Name:          name,
		ActualPrice:   actualPrice,
		DiscountPrice: discount,
	}
}
