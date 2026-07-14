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

  Order({
    required this.id,
    required this.items,
    required this.total,
    required this.status,
    required this.paymentMethod,
    required this.createdAt,
  });
}
