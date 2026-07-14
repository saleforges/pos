import 'package:flutter/foundation.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';

class AppConfig {
  static String get appName => dotenv.get('APP_NAME', fallback: 'Saleforge');
  static const String appVersion = '1.0.0';
  static const bool isProduction = kReleaseMode;
  static const bool isDevelopment = kDebugMode;
  static const String dbName = 'pos_mobile.db';
  static const int dbVersion = 1;
  static const int printerPaperWidth = 58;
  static const String printerServiceUuid = '00001101-0000-1000-8000-00805f9b34fb';
  static const Duration cacheExpiry = Duration(hours: 1);
  static const int maxCachedProducts = 1000;
}
