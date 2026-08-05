import 'package:drift/drift.dart';

@DataClassName('OrderDb')
class OrdersTable extends Table {
  TextColumn get id => text()();
  TextColumn get branchId => text()();
  TextColumn get customerId => text().nullable()();
  TextColumn get status => text()();
  TextColumn get paymentMethod => text()();
  IntColumn get total => integer()();
  IntColumn get discount => integer()();
  IntColumn get tax => integer()();
  IntColumn get serviceCharge => integer()();
  IntColumn get rounding => integer()();
  TextColumn get createdAt => text()();
  TextColumn get updatedAt => text()();

  @override
  Set<Column> get primaryKey => {id};
}
