import '../../../../../shared/models/customer.dart';
import '../remote/customer_remote_data_source.dart';
import '../local/customer_local_data_source.dart';

class CustomerRepository {
  final CustomerRemoteDataSource _remoteDataSource;
  final CustomerLocalDataSource _localDataSource;

  CustomerRepository({
    required CustomerRemoteDataSource remoteDataSource,
    required CustomerLocalDataSource localDataSource,
  })  : _remoteDataSource = remoteDataSource,
        _localDataSource = localDataSource;

  Future<List<Customer>> getCustomers() async {
    try {
      final customers = await _remoteDataSource.getCustomers();
      await _localDataSource.cacheCustomers(customers);
      return customers;
    } catch (_) {
      return await _localDataSource.getCachedCustomers();
    }
  }

  Future<Customer> createCustomer(String name, String phone) async {
    final customer = await _remoteDataSource.createCustomer(name, phone);
    await _localDataSource.addCustomer(customer);
    return customer;
  }
}
