import 'package:dio/dio.dart';
import '../../../core/config/api_config.dart';
import '../../../core/network/api_client.dart';
import '../../../shared/models/auth_response.dart';
import '../../../shared/models/user.dart';

class AuthRepository {
  final ApiClient _apiClient;

  AuthRepository(this._apiClient);

  Future<LoginResponse> login(String username, String password) async {
    try {
      final response = await _apiClient.dio.post(
        '${ApiConfig.authEndpoint}/login',
        data: LoginRequest(username: username, password: password).toJson(),
        options: Options(headers: {'Authorization': null}),
      );

      final apiResponse = ApiResponse.fromJson(response.data, (data) {
        return LoginResponse.fromJson(data as Map<String, dynamic>);
      });

      if (apiResponse.data == null) {
        throw Exception(apiResponse.message);
      }

      return apiResponse.data!;
    } on DioException catch (e) {
      if (e.response?.data != null) {
        throw Exception(e.response!.data['message'] ?? 'Login failed');
      }
      throw Exception('Network error: ${e.message}');
    }
  }

  Future<User> getCurrentUser() async {
    try {
      final response = await _apiClient.dio.get(
        '${ApiConfig.authEndpoint}/me',
      );

      final apiResponse = ApiResponse.fromJson(response.data, (data) {
        return User.fromJson(data as Map<String, dynamic>);
      });

      if (apiResponse.data == null) {
        throw Exception(apiResponse.message);
      }

      return apiResponse.data!;
    } on DioException catch (e) {
      if (e.response?.data != null) {
        throw Exception(e.response!.data['message'] ?? 'Failed to get user');
      }
      throw Exception('Network error: ${e.message}');
    }
  }

  Future<void> logout() async {
    try {
      await _apiClient.dio.post('${ApiConfig.authEndpoint}/logout');
    } catch (e) {
      // Ignore logout errors
    } finally {
      await _apiClient.clearTokens();
    }
  }

  Future<void> saveTokens(String accessToken, String refreshToken) async {
    await _apiClient.saveTokens(accessToken, refreshToken);
  }

  Future<bool> isLoggedIn() async {
    final token = await _apiClient.getAccessToken();
    return token != null;
  }

  Future<void> saveSelectedBranchId(int branchId) async {
    await _apiClient.saveSelectedBranchId(branchId);
  }

  Future<int?> getSelectedBranchId() async {
    return await _apiClient.getSelectedBranchId();
  }

  Future<void> clearSelectedBranchId() async {
    await _apiClient.clearSelectedBranchId();
  }
}
