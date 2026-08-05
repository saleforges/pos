class OrderItem {
  final String name;
  final int price;
  final int quantity;

  OrderItem({required this.name, required this.price, required this.quantity});
}

enum OrderStatus { paid, unpaid }

class Order {
  final String id;
  final List<OrderItem> items;
  final int total;
  final OrderStatus status;
  final String paymentMethod;
  final DateTime createdAt;
  final int discount;
  final int tax;
  final int serviceCharge;
  final int rounding;

  Order({
    required this.id,
    required this.items,
    required this.total,
    required this.status,
    required this.paymentMethod,
    required this.createdAt,
    this.discount = 0,
    this.tax = 0,
    this.serviceCharge = 0,
    this.rounding = 0,
  });
}
