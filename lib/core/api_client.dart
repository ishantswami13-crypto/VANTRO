import 'package:dio/dio.dart';

class ApiClient {
  final Dio dio;
  static const String baseUrl = String.fromEnvironment('API_BASE', defaultValue: 'https://vantro.onrender.com');
  static const String apiKey = String.fromEnvironment('API_KEY', defaultValue: 'supersecretapikey');

  ApiClient() : dio = Dio(BaseOptions(
    baseUrl: baseUrl,
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': apiKey,
    },
    connectTimeout: const Duration(seconds: 8),
    receiveTimeout: const Duration(seconds: 8),
  ));
}
