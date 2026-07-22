import 'dart:io';
import 'package:drift/drift.dart';
import 'package:drift/native.dart';
import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';
import 'tables/products_table.dart';
import 'tables/categories_table.dart';
import 'tables/customers_table.dart';
import 'tables/cart_table.dart';
import 'tables/cart_items_table.dart';
import 'tables/orders_table.dart';
import 'tables/order_items_table.dart';
import 'tables/settings_table.dart';
import 'tables/session_table.dart';
import 'tables/sync_queue_table.dart';

part 'app_database.g.dart';

@DriftDatabase(
  tables: [
    ProductsTable,
    CategoriesTable,
    CustomersTable,
    CartTable,
    CartItemsTable,
    OrdersTable,
    OrderItemsTable,
    SettingsTable,
    SessionTable,
    SyncQueueTable,
  ],
)
class AppDatabase extends _$AppDatabase {
  AppDatabase() : super(_openConnection());

  @override
  int get schemaVersion => 1;

  static QueryExecutor _openConnection() {
    return LazyDatabase(() async {
      final dir = await getApplicationDocumentsDirectory();
      await dir.create(recursive: true);
      final file = File(p.join(dir.path, 'pos_mobile.db'));
      return NativeDatabase(file);
    });
  }
}
