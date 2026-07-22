import '../../../../../shared/models/order.dart';

class OrderLocalDataSource {
  Future<void> saveOrder(Order order) async {
    // TODO: Implement local DB storage
  }

  Future<List<Order>> getOrders() async {
    // TODO: Implement local DB retrieval
    return [];
  }

  Future<void> updateOrderStatus(String orderId, String status) async {
    // TODO: Implement local DB update
  }
}
