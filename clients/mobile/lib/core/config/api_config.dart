class ApiConfig {
  static const String baseUrl = 'https://api-dev.saleforges.com';
  static const String apiVersion = 'v1';
  static const String authEndpoint = '/api/$apiVersion/auth';
  static const String productsEndpoint = '/api/$apiVersion/products';
  static const String categoriesEndpoint = '/api/$apiVersion/categories';
  static const String transactionsEndpoint = '/api/$apiVersion/transactions';
  static const String syncEndpoint = '/api/$apiVersion/sync';
  static const int connectTimeout = 30000;
  static const int receiveTimeout = 30000;
  static const int syncIntervalSeconds = 30;
}
