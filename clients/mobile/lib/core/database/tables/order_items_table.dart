import 'package:drift/drift.dart';

@DataClassName('OrderItemDb')
class OrderItemsTable extends Table {
  TextColumn get id => text()();
  TextColumn get orderId => text()();
  TextColumn get productId => text()();
  TextColumn get name => text()();
  IntColumn get price => integer()();
  IntColumn get quantity => integer()();

  @override
  Set<Column> get primaryKey => {id};
}
