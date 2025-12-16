package basket

type BasketProduct struct {
	ProductID string
	Count     int
}

func NewBasketProduct(id string, count int) BasketProduct {
	return BasketProduct{
		ProductID: id,
		Count:     count,
	}
}
