import 'package:drift/drift.dart';

@DataClassName('ProductDb')
class ProductsTable extends Table {
  TextColumn get id => text()();
  TextColumn get name => text()();
  IntColumn get price => integer()();
  IntColumn get stock => integer()();
  TextColumn get category => text()();
  TextColumn? get image => text().nullable()();
  TextColumn? get subtitle => text().nullable()();
  BoolColumn get isFavorite => boolean()();
  BoolColumn get hasVariants => boolean()();
  TextColumn get createdAt => text()();
  TextColumn get updatedAt => text()();

  @override
  Set<Column> get primaryKey => {id};
}
