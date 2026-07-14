import 'package:flutter_dotenv/flutter_dotenv.dart';

class ApiConfig {
  static String get baseUrl => dotenv.get('BASE_URL', fallback: 'https://api-dev.saleforges.com');
  static const String apiVersion = 'v1';
  static const String authEndpoint = '/$apiVersion/auth';
  static const String productsEndpoint = '/$apiVersion/products';
  static const String categoriesEndpoint = '/$apiVersion/categories';
  static const String transactionsEndpoint = '/$apiVersion/transactions';
  static const String syncEndpoint = '/$apiVersion/sync';
  static const int connectTimeout = 30000;
  static const int receiveTimeout = 30000;
  static const int syncIntervalSeconds = 30;
}
