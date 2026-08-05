import '../../../../../shared/models/order.dart';
import '../remote/order_remote_data_source.dart';
import '../local/order_local_data_source.dart';

class OrderRepository {
  final OrderRemoteDataSource _remoteDataSource;
  final OrderLocalDataSource _localDataSource;

  OrderRepository({
    required OrderRemoteDataSource remoteDataSource,
    required OrderLocalDataSource localDataSource,
  })  : _remoteDataSource = remoteDataSource,
        _localDataSource = localDataSource;

  Future<void> createOrder(Order order) async {
    await _localDataSource.saveOrder(order);
    try {
      await _remoteDataSource.createOrder(order);
    } catch (_) {
      // Will be synced later
    }
  }

  Future<List<Order>> getOrders() async {
    try {
      return await _remoteDataSource.getOrders();
    } catch (_) {
      return await _localDataSource.getOrders();
    }
  }

  Future<void> markPaid(String orderId) async {
    await _localDataSource.updateOrderStatus(orderId, 'paid');
    try {
      await _remoteDataSource.syncOrder(
        Order(
          id: orderId,
          items: [],
          total: 0,
          status: OrderStatus.paid,
          paymentMethod: '',
          createdAt: DateTime.now(),
        ),
      );
    } catch (_) {
      // Will be synced later
    }
  }
}
