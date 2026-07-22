import 'package:drift/drift.dart';
import '../app_database.dart';
import '../../../shared/models/product.dart';

class ProductDao {
  final AppDatabase _db;
  final _batchSize = 50;

  ProductDao(this._db);

  Future<List<Product>> getAllProducts() async {
    final rows = await _db.select(_db.productsTable).get();
    return rows.map(_toProduct).toList();
  }

  Future<bool> hasProducts() async {
    final rows = await _db.select(_db.productsTable).get();
    return rows.isNotEmpty;
  }

  Future<void> upsertProducts(List<Product> products) async {
    for (var i = 0; i < products.length; i += _batchSize) {
      final batch = products.skip(i).take(_batchSize).toList();
      await _db.batch((batchDb) {
        for (final product in batch) {
          batchDb.insert(
            _db.productsTable,
            _toProductDb(product),
            mode: InsertMode.insertOrReplace,
          );
        }
      });
    }
  }

  Future<void> clearProducts() async {
    await _db.delete(_db.productsTable).go();
  }

  Product _toProduct(ProductDb row) {
    return Product(
      id: row.id,
      name: row.name,
      sku: row.sku,
      barcode: row.barcode,
      description: row.description,
      sellingPrice: row.sellingPrice,
      costPrice: row.costPrice,
      stock: row.stock,
      category: row.category,
      image: row.image,
      subtitle: row.subtitle,
      unit: row.unit,
      isActive: row.isActive,
      isFavorite: row.isFavorite,
      hasVariants: row.hasVariants,
      createdAt: DateTime.parse(row.createdAt),
      updatedAt: DateTime.parse(row.updatedAt),
    );
  }

  ProductsTableCompanion _toProductDb(Product product) {
    return ProductsTableCompanion(
      id: Value(product.id),
      name: Value(product.name),
      sku: Value(product.sku),
      barcode: Value(product.barcode),
      description: Value(product.description),
      sellingPrice: Value(product.sellingPrice),
      costPrice: Value(product.costPrice),
      stock: Value(product.stock),
      category: Value(product.category),
      image: Value(product.image),
      subtitle: Value(product.subtitle),
      unit: Value(product.unit),
      isActive: Value(product.isActive),
      isFavorite: Value(product.isFavorite),
      hasVariants: Value(product.hasVariants),
      createdAt: Value(product.createdAt.toIso8601String()),
      updatedAt: Value(product.updatedAt.toIso8601String()),
    );
  }
}
