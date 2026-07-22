import '../../../../../shared/models/product.dart';

class ProductLocalDataSource {
  Future<void> cacheProducts(List<Product> products) async {
    // TODO: Implement local DB storage
  }

  Future<List<Product>> getCachedProducts() async {
    // TODO: Implement local DB retrieval
    return [];
  }

  Future<void> clearCache() async {
    // TODO: Implement cache clearing
  }
}
