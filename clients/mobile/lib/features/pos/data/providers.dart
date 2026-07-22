import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'remote/product_remote_data_source.dart';
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

// Remote Data Sources
final productRemoteDataSourceProvider = Provider<ProductRemoteDataSource>((ref) {
  return ProductRemoteDataSource();
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
  return ProductLocalDataSource();
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
