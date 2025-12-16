package product

import "context"

func (s *Services) GetProductDetails(ctx context.Context, id string) (ProductDetail, error) {
	return NewProductDetail(id,
		"essa",
		[]string{"/public/products/essa/1.webp", "/public/products/essa/2.webp", "/public/products/essa/3.webp"},
		[]string{
			`Nowa podkaszarka Daewoo. 
				Dzięki niskiej wadze i niedużym rozmiarom 
				podkaszarka DATR 800E świetnie sprawdzi się na małej działce czy w ogródku przydomowym.
				`,
			`Nowa podkaszarka Daewoo. 
				Dzięki niskiej wadze i niedużym rozmiarom 
				podkaszarka DATR 800E świetnie sprawdzi się na małej działce czy w ogródku przydomowym.
				`,
			`Nowa podkaszarka Daewoo. 
				Dzięki niskiej wadze i niedużym rozmiarom 
				podkaszarka DATR 800E świetnie sprawdzi się na małej działce czy w ogródku przydomowym.
				`,
		},
		[]string{},
		1.99), nil
}
