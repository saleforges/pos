import '../../../../shared/models/auth_response.dart';
import '../../../../shared/models/user.dart';
import '../../../core/storage/session_local_data_source.dart';
import 'remote/auth_remote_data_source.dart';
import 'local/auth_local_data_source.dart';

class AuthRepository {
  final AuthRemoteDataSource _remoteDataSource;
  final AuthLocalDataSource _localDataSource;
  final SessionLocalDataSource _sessionDataSource;

  AuthRepository({
    required AuthRemoteDataSource remoteDataSource,
    required AuthLocalDataSource localDataSource,
    required SessionLocalDataSource sessionDataSource,
  })  : _remoteDataSource = remoteDataSource,
        _localDataSource = localDataSource,
        _sessionDataSource = sessionDataSource;

  Future<LoginResponse> login(String username, String password) async {
    final response = await _remoteDataSource.login(username, password);
    await _sessionDataSource.saveAccessToken(response.accessToken);
    await _sessionDataSource.saveRefreshToken(response.refreshToken);
    return response;
  }

  Future<User> getCurrentUser() async {
    final user = await _remoteDataSource.getCurrentUser();
    await _localDataSource.cacheUser(user);
    return user;
  }

  Future<void> logout() async {
    await _remoteDataSource.logout();
    await _sessionDataSource.clearAccessToken();
    await _sessionDataSource.clearRefreshToken();
    await _sessionDataSource.clearSelectedBranchId();
    await _localDataSource.clearCachedUser();
  }

  Future<bool> isLoggedIn() async {
    return await _sessionDataSource.isLoggedIn();
  }

  Future<void> saveTokens(String accessToken, String refreshToken) async {
    await _sessionDataSource.saveAccessToken(accessToken);
    await _sessionDataSource.saveRefreshToken(refreshToken);
  }

  Future<void> saveSelectedBranchId(int branchId) async {
    await _sessionDataSource.saveSelectedBranchId(branchId);
  }

  Future<int?> getSelectedBranchId() async {
    return await _sessionDataSource.getSelectedBranchId();
  }

  Future<void> clearSelectedBranchId() async {
    await _sessionDataSource.clearSelectedBranchId();
  }
}
