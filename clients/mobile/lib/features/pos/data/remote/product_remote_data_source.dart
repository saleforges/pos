import '../../../../../shared/models/product.dart';

class ProductRemoteDataSource {
  Future<List<Product>> getProducts({String? category, String? searchQuery}) async {
    // TODO: Implement API call
    throw UnimplementedError('ProductRemoteDataSource.getProducts not yet implemented');
  }

  Future<Product> getProduct(String id) async {
    throw UnimplementedError('ProductRemoteDataSource.getProduct not yet implemented');
  }
}
