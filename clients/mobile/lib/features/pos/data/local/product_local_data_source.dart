import '../../../../../shared/models/product.dart';
import '../../../../core/database/daos/product_dao.dart';

class ProductLocalDataSource {
  final ProductDao _dao;

  ProductLocalDataSource(this._dao);

  Future<void> cacheProducts(List<Product> products) async {
    await _dao.upsertProducts(products);
  }

  Future<List<Product>> getCachedProducts() async {
    return await _dao.getAllProducts();
  }

  Future<bool> hasCachedProducts() async {
    return await _dao.hasProducts();
  }

  Future<void> clearCache() async {
    await _dao.clearProducts();
  }
}
