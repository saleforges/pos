import 'package:drift/drift.dart';

@DataClassName('SessionDb')
class SessionTable extends Table {
  TextColumn get key => text()();
  TextColumn get value => text()();

  @override
  Set<Column> get primaryKey => {key};
}
