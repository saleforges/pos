import 'package:drift/drift.dart';

@DataClassName('ProductDb')
class ProductsTable extends Table {
  TextColumn get id => text()();
  TextColumn get name => text()();
  TextColumn get sku => text()();
  TextColumn? get barcode => text().nullable()();
  TextColumn? get description => text().nullable()();
  IntColumn get sellingPrice => integer()();
  IntColumn get costPrice => integer()();
  IntColumn get stock => integer()();
  TextColumn get category => text()();
  TextColumn? get image => text().nullable()();
  TextColumn? get subtitle => text().nullable()();
  TextColumn get unit => text()();
  BoolColumn get isActive => boolean()();
  BoolColumn get isFavorite => boolean()();
  BoolColumn get hasVariants => boolean()();
  TextColumn get createdAt => text()();
  TextColumn get updatedAt => text()();

  @override
  Set<Column> get primaryKey => {id};
}
