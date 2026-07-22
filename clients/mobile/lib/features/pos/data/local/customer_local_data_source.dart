import '../../../../../shared/models/customer.dart';

class CustomerLocalDataSource {
  Future<void> cacheCustomers(List<Customer> customers) async {
    // TODO: Implement local DB storage
  }

  Future<List<Customer>> getCachedCustomers() async {
    // TODO: Implement local DB retrieval
    return [];
  }

  Future<void> addCustomer(Customer customer) async {
    // TODO: Implement local DB insert
  }
}
