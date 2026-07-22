import 'package:drift/drift.dart';

@DataClassName('CartItemDb')
class CartItemsTable extends Table {
  TextColumn get id => text()();
  TextColumn get cartId => text()();
  TextColumn get productId => text()();
  IntColumn get quantity => integer()();
  IntColumn get price => integer()();
  TextColumn get notes => text()();

  @override
  Set<Column> get primaryKey => {id};
}
