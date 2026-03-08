package product

import "context"

type ProductDetailsQuery struct {
	ID string
}

func (s *Services) getProductDetails(ctx context.Context, query ProductDetailsQuery) (ProductDetail, error) {
	return NewProductDetail(query.ID,
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
