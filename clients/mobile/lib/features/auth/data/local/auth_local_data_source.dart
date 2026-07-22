import '../../../../shared/models/user.dart';

class AuthLocalDataSource {
  Future<void> cacheUser(User user) async {
    // TODO: Implement for offline support
  }

  Future<User?> getCachedUser() async {
    // TODO: Implement for offline support
    return null;
  }

  Future<void> clearCachedUser() async {
    // TODO: Implement for offline support
  }
}
