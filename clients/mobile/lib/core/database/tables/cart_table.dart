import 'package:drift/drift.dart';

@DataClassName('CartDb')
class CartTable extends Table {
  TextColumn get id => text()();
  TextColumn get branchId => text()();
  TextColumn get customerId => text().nullable()();
  TextColumn get status => text()();
  TextColumn get createdAt => text()();
  TextColumn get updatedAt => text()();

  @override
  Set<Column> get primaryKey => {id};
}
