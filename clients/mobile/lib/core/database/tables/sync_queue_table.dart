import 'package:drift/drift.dart';

@DataClassName('SyncQueueDb')
class SyncQueueTable extends Table {
  TextColumn get id => text()();
  TextColumn get entityType => text()();
  TextColumn get entityId => text()();
  TextColumn get operation => text()();
  TextColumn get payload => text()();
  TextColumn get status => text()();
  IntColumn get retryCount => integer()();
  TextColumn get createdAt => text()();
  TextColumn? get lastAttemptAt => text().nullable()();

  @override
  Set<Column> get primaryKey => {id};
}
