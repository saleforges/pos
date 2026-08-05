import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class SessionLocalDataSource {
  final FlutterSecureStorage _storage;

  SessionLocalDataSource({FlutterSecureStorage? storage})
      : _storage = storage ?? const FlutterSecureStorage();

  Future<void> saveAccessToken(String token) async {
    await _storage.write(key: 'access_token', value: token);
  }

  Future<String?> getAccessToken() async {
    return await _storage.read(key: 'access_token');
  }

  Future<void> clearAccessToken() async {
    await _storage.delete(key: 'access_token');
  }

  Future<void> saveRefreshToken(String token) async {
    await _storage.write(key: 'refresh_token', value: token);
  }

  Future<String?> getRefreshToken() async {
    return await _storage.read(key: 'refresh_token');
  }

  Future<void> clearRefreshToken() async {
    await _storage.delete(key: 'refresh_token');
  }

  Future<void> saveSelectedBranchId(int branchId) async {
    await _storage.write(key: 'selected_branch_id', value: branchId.toString());
  }

  Future<int?> getSelectedBranchId() async {
    final value = await _storage.read(key: 'selected_branch_id');
    return value != null ? int.tryParse(value) : null;
  }

  Future<void> clearSelectedBranchId() async {
    await _storage.delete(key: 'selected_branch_id');
  }

  Future<void> saveLocale(String locale) async {
    await _storage.write(key: 'locale', value: locale);
  }

  Future<String?> getLocale() async {
    return await _storage.read(key: 'locale');
  }

  Future<void> clearAll() async {
    await _storage.deleteAll();
  }

  Future<bool> isLoggedIn() async {
    final token = await getAccessToken();
    return token != null;
  }
}
