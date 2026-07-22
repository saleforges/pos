import '../remote/category_remote_data_source.dart';
import '../local/category_local_data_source.dart';

class CategoryRepository {
  final CategoryRemoteDataSource _remoteDataSource;
  final CategoryLocalDataSource _localDataSource;

  CategoryRepository({
    required CategoryRemoteDataSource remoteDataSource,
    required CategoryLocalDataSource localDataSource,
  })  : _remoteDataSource = remoteDataSource,
        _localDataSource = localDataSource;

  Future<List<String>> getCategories() async {
    try {
      final categories = await _remoteDataSource.getCategories();
      await _localDataSource.cacheCategories(categories);
      return categories;
    } catch (_) {
      return await _localDataSource.getCachedCategories();
    }
  }
}
