import '../cart_store.dart';

class CartLocalDataSource {
  Future<void> saveCart(List<CartItem> items) async {
    // TODO: Implement local DB storage
  }

  Future<List<CartItem>> loadCart() async {
    // TODO: Implement local DB retrieval
    return [];
  }

  Future<void> clearCart() async {
    // TODO: Implement local DB clearing
  }
}
