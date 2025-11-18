import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/providers.dart';
import '../../core/models/coach.dart';

class CoachScreen extends ConsumerStatefulWidget {
  const CoachScreen({super.key});
  @override
  ConsumerState<CoachScreen> createState() => _CoachScreenState();
}

class _CoachScreenState extends ConsumerState<CoachScreen> {
  CoachPlan? _plan; bool _loading=false; String? _error;
  final _incomeCtrl = TextEditingController(text: '60000');
  final _rentCtrl = TextEditingController(text: '15000');
  final _goalCtrl = TextEditingController(text: 'iPhone 16');

  Future<void> _generate() async {
    setState(()=>{_loading=true,_error=null});
    final dio = ref.read(apiProvider);
    try {
      final resp = await dio.post('/api/coach/plan', data: {
        'income_cents': (int.tryParse(_incomeCtrl.text.trim()) ?? 0)*100,
        'rent_cents': (int.tryParse(_rentCtrl.text.trim()) ?? 0)*100,
        'goal': _goalCtrl.text.trim(),
      });
      setState(()=>_plan = CoachPlan.fromJson(Map<String,dynamic>.from(resp.data)));
    } on DioException catch (e) { setState(()=>_error = e.response?.data?.toString() ?? e.message);
    } catch (e) { setState(()=>_error = e.toString());
    } finally { setState(()=>_loading=false); }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Money Coach')),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(children: [
          Row(children: [
            Expanded(child: TextField(controller: _incomeCtrl, keyboardType: TextInputType.number, decoration: const InputDecoration(labelText:'Income (₹ / month)', border: OutlineInputBorder()))),
            const SizedBox(width: 8),
            Expanded(child: TextField(controller: _rentCtrl, keyboardType: TextInputType.number, decoration: const InputDecoration(labelText:'Rent (₹ / month)', border: OutlineInputBorder()))),
          ]),
          const SizedBox(height: 8),
          TextField(controller: _goalCtrl, decoration: const InputDecoration(labelText:'Goal (pot to fund)', border: OutlineInputBorder())),
          const SizedBox(height: 12),
          SizedBox(width: double.infinity, child: FilledButton(onPressed: _loading?null:_generate, child: _loading?const SizedBox(height:18,width:18,child:CircularProgressIndicator(strokeWidth:2)):const Text('Generate Plan'))),
          const SizedBox(height: 16),
          if (_error != null) Text(_error!, style: const TextStyle(color: Colors.red)),
          if (_plan != null) _PlanCard(plan: _plan!),
        ]),
      ),
    );
  }
}

class _PlanCard extends StatelessWidget {
  final CoachPlan plan;
  const _PlanCard({required this.plan});
  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity, padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(16),
        boxShadow: const [BoxShadow(color: Color(0x11000000), blurRadius: 16, offset: Offset(0,8))]),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Text('Week starting ${plan.weekStart}', style: const TextStyle(fontWeight: FontWeight.w700)),
        const SizedBox(height: 8),
        Text('Health Score: ${plan.healthScore}', style: const TextStyle(fontWeight: FontWeight.w600)),
        const SizedBox(height: 8),
        const Text('Rules'),
        const SizedBox(height: 6),
        ...plan.rules.map((r)=>Row(children:[const Text('• '), Expanded(child: Text(r))])),
        const SizedBox(height: 8),
        Text('Nudge: ${plan.dailyNudge}', style: const TextStyle(color: Colors.black54)),
      ]),
    );
  }
}
