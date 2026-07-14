import '../../../../shared/models/order.dart';

class OrderStore {
  final List<Order> _orders = [];
  int _nextId = 1;

  List<Order> get orders => List.unmodifiable(_orders);

  List<Order> get unpaidOrders => _orders.where((o) => o.status == OrderStatus.unpaid).toList();

  void add(Order order) {
    _orders.insert(0, order);
  }

  void markPaid(String id) {
    final i = _orders.indexWhere((o) => o.id == id);
    if (i != -1) {
      _orders[i] = Order(
        id: _orders[i].id,
        items: _orders[i].items,
        total: _orders[i].total,
        status: OrderStatus.paid,
        paymentMethod: _orders[i].paymentMethod,
        createdAt: _orders[i].createdAt,
      );
    }
  }

  String nextId() => 'ORD-${_nextId++}';
}
