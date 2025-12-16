package product

import "context"

func (s *Services) GetPromotions(ctx context.Context) ([]Product, error) {
	return []Product{
		NewPromoProduct("1", "Podkaszarka elektryczna DAEWOO DATR 800E 550W", "/public/products/essa/1.webp", 199.99, 112.22),
		NewPromoProduct("2", "Podkaszarka elektryczna DAEWOO DATR 800E 550W", "/public/products/essa/1.webp", 233.99, 199.00),
		NewPromoProduct("3", "Podkaszarka elektryczna DAEWOO DATR 800E 550W", "/public/products/essa/1.webp", 342.32, 300.20),
		NewPromoProduct("4", "Podkaszarka elektryczna DAEWOO DATR 800E 550W", "/public/products/essa/1.webp", 499.99, 99.00),
	}, nil
}
