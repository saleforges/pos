import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../shared/models/product.dart';
import '../../../../core/database/app_database.dart';
import '../../../../core/database/daos/product_dao.dart';
import 'remote/product_remote_data_source.dart';
import 'remote/fake_product_remote_data_source.dart';
import 'remote/category_remote_data_source.dart';
import 'remote/customer_remote_data_source.dart';
import 'remote/order_remote_data_source.dart';
import 'local/product_local_data_source.dart';
import 'local/category_local_data_source.dart';
import 'local/customer_local_data_source.dart';
import 'local/order_local_data_source.dart';
import 'local/cart_local_data_source.dart';
import 'repositories/product_repository.dart';
import 'repositories/category_repository.dart';
import 'repositories/customer_repository.dart';
import 'repositories/order_repository.dart';
import 'product_list_notifier.dart';

final appDatabaseProvider = Provider<AppDatabase>((ref) {
  return AppDatabase();
});

final productDaoProvider = Provider<ProductDao>((ref) {
  return ref.read(appDatabaseProvider).productDao;
});

// Remote Data Sources
final productRemoteDataSourceProvider = Provider<ProductRemoteDataSource>((ref) {
  return FakeProductRemoteDataSource();
});

final categoryRemoteDataSourceProvider = Provider<CategoryRemoteDataSource>((ref) {
  return CategoryRemoteDataSource();
});

final customerRemoteDataSourceProvider = Provider<CustomerRemoteDataSource>((ref) {
  return CustomerRemoteDataSource();
});

final orderRemoteDataSourceProvider = Provider<OrderRemoteDataSource>((ref) {
  return OrderRemoteDataSource();
});

// Local Data Sources
final productLocalDataSourceProvider = Provider<ProductLocalDataSource>((ref) {
  return ProductLocalDataSource(ref.read(productDaoProvider));
});

final categoryLocalDataSourceProvider = Provider<CategoryLocalDataSource>((ref) {
  return CategoryLocalDataSource();
});

final customerLocalDataSourceProvider = Provider<CustomerLocalDataSource>((ref) {
  return CustomerLocalDataSource();
});

final orderLocalDataSourceProvider = Provider<OrderLocalDataSource>((ref) {
  return OrderLocalDataSource();
});

final cartLocalDataSourceProvider = Provider<CartLocalDataSource>((ref) {
  return CartLocalDataSource();
});

// Repositories
final productRepositoryProvider = Provider<ProductRepository>((ref) {
  return ProductRepository(
    remoteDataSource: ref.read(productRemoteDataSourceProvider),
    localDataSource: ref.read(productLocalDataSourceProvider),
  );
});

final categoryRepositoryProvider = Provider<CategoryRepository>((ref) {
  return CategoryRepository(
    remoteDataSource: ref.read(categoryRemoteDataSourceProvider),
    localDataSource: ref.read(categoryLocalDataSourceProvider),
  );
});

final customerRepositoryProvider = Provider<CustomerRepository>((ref) {
  return CustomerRepository(
    remoteDataSource: ref.read(customerRemoteDataSourceProvider),
    localDataSource: ref.read(customerLocalDataSourceProvider),
  );
});

final orderRepositoryProvider = Provider<OrderRepository>((ref) {
  return OrderRepository(
    remoteDataSource: ref.read(orderRemoteDataSourceProvider),
    localDataSource: ref.read(orderLocalDataSourceProvider),
  );
});

final productListProvider = StateNotifierProvider<ProductListNotifier, List<Product>>((ref) {
  return ProductListNotifier(ref.read(productRepositoryProvider));
});
