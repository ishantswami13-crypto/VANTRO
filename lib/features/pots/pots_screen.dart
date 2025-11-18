import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/providers.dart';
import '../../core/models/pot.dart';

class PotsScreen extends ConsumerStatefulWidget {
  const PotsScreen({super.key});
  @override
  ConsumerState<PotsScreen> createState() => _PotsScreenState();
}

class _PotsScreenState extends ConsumerState<PotsScreen> {
  List<SavingPot> _items = [];
  bool _loading = false; String? _error;
  final _nameCtrl = TextEditingController(); final _targetCtrl = TextEditingController();

  @override
  void initState() { super.initState(); _fetch(); }

  Future<void> _fetch() async {
    setState(()=>{_loading=true,_error=null});
    final dio = ref.read(apiProvider);
    try {
      final resp = await dio.get('/api/pots');
      final data = resp.data as List? ?? [];
      setState(()=>_items = data.map((e)=>SavingPot.fromJson(Map<String,dynamic>.from(e))).toList());
    } on DioException catch (e) { setState(()=>_error = e.response?.data?.toString() ?? e.message);
    } catch (e) { setState(()=>_error = e.toString());
    } finally { setState(()=>_loading=false); }
  }

  Future<void> _create() async {
    final name = _nameCtrl.text.trim();
    final target = int.tryParse(_targetCtrl.text.trim()) ?? 0;
    if (name.isEmpty || target <= 0) return;
    final dio = ref.read(apiProvider);
    try {
      await dio.post('/api/pots', data: {'name':name,'target_cents':target*100});
      _nameCtrl.clear(); _targetCtrl.clear(); _fetch();
    } catch (_) {}
  }

  Future<void> _add(String id, int rupees) async {
    final dio = ref.read(apiProvider);
    try { await dio.patch('/api/pots/$id', data: {'add_cents': rupees*100}); _fetch(); } catch (_) {}
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Saving Pots')),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: _loading? const Center(child:CircularProgressIndicator())
          : _error!=null? Center(child: Text(_error!))
          : Column(children: [
              Row(children: [
                Expanded(child: TextField(controller: _nameCtrl, decoration: const InputDecoration(labelText:'Pot name', border: OutlineInputBorder()))),
                const SizedBox(width: 8),
                SizedBox(width: 140, child: TextField(controller: _targetCtrl, keyboardType: TextInputType.number, decoration: const InputDecoration(labelText:'Target ₹', border: OutlineInputBorder()))),
                const SizedBox(width: 8),
                FilledButton(onPressed: _create, child: const Text('Create')),
              ]),
              const SizedBox(height: 12),
              Expanded(child: ListView.separated(
                itemCount: _items.length, separatorBuilder: (_, __)=>const SizedBox(height:10),
                itemBuilder: (_, i){
                  final p = _items[i];
                  final saved = (p.savedCents/100).toStringAsFixed(2);
                  final target = (p.targetCents/100).toStringAsFixed(2);
                  return Container(
                    padding: const EdgeInsets.all(14),
                    decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(14),
                      boxShadow: const [BoxShadow(color: Color(0x0F000000), blurRadius: 12, offset: Offset(0,6))]),
                    child: Row(children: [
                      Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                        Text(p.name, style: const TextStyle(fontWeight: FontWeight.w600)),
                        const SizedBox(height: 6),
                        LinearProgressIndicator(value: p.progress.clamp(0, 1)),
                        const SizedBox(height: 6),
                        Text('₹$saved / ₹$target', style: const TextStyle(color: Colors.black54)),
                      ])),
                      const SizedBox(width: 8),
                      OutlinedButton(onPressed: ()=>_add(p.id, 500), child: const Text('+ ₹500')),
                    ]),
                  );
                },
              )),
            ]),
      ),
    );
  }
}
