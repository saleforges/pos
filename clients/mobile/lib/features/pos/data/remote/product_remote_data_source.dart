import '../../../../../shared/models/product.dart';

class ProductRemoteDataSource {
  Future<List<Product>> getProducts() async {
    throw UnimplementedError('ProductRemoteDataSource.getProducts not yet implemented');
  }

  Future<Product> getProduct(String id) async {
    throw UnimplementedError('ProductRemoteDataSource.getProduct not yet implemented');
  }
}
