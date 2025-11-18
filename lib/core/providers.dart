import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'api_client.dart';

// single Dio client
final apiProvider = Provider((ref) => ApiClient().dio);
