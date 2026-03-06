package product

type productCategory string

var (
	MachinesProductCategory        productCategory = "MACHINES"
	GardeningProductCategory       productCategory = "GARDENING"
	PartsProductCategory           productCategory = "PARTS"
	ElectroProductCategory         productCategory = "ELECTRO"
	ElectroMachinesProductCategory                 = "ELECTROMACHINES"
)

type Product struct {
	ID          string
	Name        string
	Image       string
	PriceBefore int
	Price       int
	Category    productCategory
}

type SortSpecifier string

const (
	PriceAsc  SortSpecifier = "price-asc"
	PriceDesc SortSpecifier = "price-desc"
	NameAsc   SortSpecifier = "name-asc"
	NameDesc  SortSpecifier = "name-desc"
)

func SpecifySort(candidate string) SortSpecifier {
	switch candidate {
	case string(PriceAsc):
		return PriceAsc
	case string(PriceDesc):
		return PriceDesc
	case string(NameDesc):
		return NameDesc
	default:
		return NameAsc
	}
}

type Filter struct {
	PriceFrom              int
	PriceTo                int
	IncludeMachines        bool
	IncludeGardening       bool
	IncludeParts           bool
	IncludeElectro         bool
	IncludeElectroMachines bool
	Sort                   SortSpecifier
	Search                 string
}

func NewFilter(
	priceFrom, priceTo int,
	includeMachines, includeGardening, includeParts, includeElectro, includeElectroMachines bool,
	sort, search string,
) Filter {
	return Filter{
		PriceFrom:              priceFrom,
		PriceTo:                priceTo,
		IncludeMachines:        includeMachines,
		IncludeGardening:       includeGardening,
		IncludeParts:           includeParts,
		IncludeElectro:         includeElectro,
		IncludeElectroMachines: includeElectroMachines,
		Sort:                   SpecifySort(sort),
		Search:                 search,
	}
}

type ProductDetail struct {
	ID                  string
	Name                string
	Images              []string
	PriceBefore         float32
	Price               float32
	ProductInfo         []string
	TechnicalParameters []string
}

func NewProductDetail(id, name string, images, productInfo, technicalParameters []string, price float32) ProductDetail {
	return ProductDetail{
		ID:                  id,
		Name:                name,
		Images:              images,
		ProductInfo:         productInfo,
		TechnicalParameters: technicalParameters,
		Price:               price,
	}
}

func NewPromoProductDetail(id, name string, images, productInfo, technicalParameters []string, price float32, priceBefore float32) ProductDetail {
	return ProductDetail{
		ID:                  id,
		Name:                name,
		Images:              images,
		ProductInfo:         productInfo,
		TechnicalParameters: technicalParameters,
		Price:               price,
		PriceBefore:         priceBefore,
	}
}

type BasketProduct struct {
	ProductID string
	Count     int
}

func NewProduct(id, name, image string, price int) Product {
	return Product{
		ID:    id,
		Name:  name,
		Price: price,
		Image: image,
	}
}

func (p Product) ProductPrice() int {
	if p.PriceBefore != 0 {
		return p.PriceBefore
	}
	return p.Price
}

func NewPromoProduct(id, name, image string, priceBefore int, price int) Product {
	return Product{
		ID:          id,
		Name:        name,
		Price:       price,
		Image:       image,
		PriceBefore: priceBefore,
	}
}
