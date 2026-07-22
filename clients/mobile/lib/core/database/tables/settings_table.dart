import 'package:drift/drift.dart';

@DataClassName('SettingDb')
class SettingsTable extends Table {
  TextColumn get key => text()();
  TextColumn get value => text()();

  @override
  Set<Column> get primaryKey => {key};
}
