package store_domain

type orderStatus string

var (
	OrderPaid      orderStatus = "PAID"
	OrderShipping  orderStatus = "SHIPPING"
	OrderFinalized orderStatus = "FINALIZED"
)

func OrderStatusFromRawStatus(rawStatus string) orderStatus {
	switch rawStatus {
	case string(OrderShipping):
		return OrderShipping
	case string(OrderFinalized):
		return OrderFinalized
	default:
		return OrderPaid
	}
}

type Order struct {
	ID         string
	CheckoutID string
	CreatedAt  string
	Products   []OrderProduct
	Status     orderStatus
}

type OrderProduct struct {
	Name      string
	Count     int
	ItemPrice float32
}

func NewOrder(id, checkoutID string, createdAt string, status orderStatus, products []OrderProduct) Order {
	return Order{
		ID:         id,
		CheckoutID: checkoutID,
		CreatedAt:  createdAt,
		Status:     status,
		Products:   products,
	}
}

func NewOrderProduct(name string, count int, itemPrice float32) OrderProduct {
	return OrderProduct{
		Name:      name,
		Count:     count,
		ItemPrice: itemPrice,
	}
}
