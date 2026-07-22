import '../../../../../shared/models/product.dart';
import '../remote/product_remote_data_source.dart';
import '../local/product_local_data_source.dart';

class ProductRepository {
  final ProductRemoteDataSource _remoteDataSource;
  final ProductLocalDataSource _localDataSource;

  ProductRepository({
    required ProductRemoteDataSource remoteDataSource,
    required ProductLocalDataSource localDataSource,
  })  : _remoteDataSource = remoteDataSource,
        _localDataSource = localDataSource;

  Future<List<Product>> getCachedProducts() async {
    return await _localDataSource.getCachedProducts();
  }

  Future<bool> hasCachedProducts() async {
    return await _localDataSource.hasCachedProducts();
  }

  Future<List<Product>> refreshProducts() async {
    final fresh = await _remoteDataSource.getProducts();
    await _localDataSource.cacheProducts(fresh);
    return fresh;
  }

  Future<Product> getProduct(String id) async {
    try {
      return await _remoteDataSource.getProduct(id);
    } catch (_) {
      final cached = await _localDataSource.getCachedProducts();
      return cached.firstWhere((p) => p.id == id);
    }
  }
}
