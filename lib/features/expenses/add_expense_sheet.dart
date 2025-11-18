import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/providers.dart';

class AddExpenseSheet extends ConsumerStatefulWidget {
  const AddExpenseSheet({super.key});
  @override
  ConsumerState<AddExpenseSheet> createState() => _AddExpenseSheetState();
}

class _AddExpenseSheetState extends ConsumerState<AddExpenseSheet> {
  final _amountCtrl = TextEditingController();
  final _noteCtrl = TextEditingController();
  String _category = 'food';
  String _mood = 'calm';
  bool _saving = false;
  String? _error;

  final _cats = const ['food', 'travel', 'bills', 'shopping', 'misc'];
  final _moods = const ['calm', 'normal', 'stressed'];

  Future<void> _submit() async {
    final amt = int.tryParse(_amountCtrl.text.trim());
    if (amt == null || amt <= 0) { setState(()=>_error='Enter amount in ₹ (e.g., 199)'); return; }
    setState(()=>_saving=true);
    final dio = ref.read(apiProvider);
    try {
      await dio.post('/api/expenses', data: {
        'amount_cents': amt * 100,
        'category': _category,
        'mood': _mood,
        'note': _noteCtrl.text.trim(),
      });
      if (mounted) Navigator.of(context).pop(true);
    } on DioException catch (e) {
      setState(()=>_error = e.response?.data?.toString() ?? e.message);
    } catch (e) {
      setState(()=>_error = e.toString());
    } finally {
      setState(()=>_saving=false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final bottom = MediaQuery.of(context).viewInsets.bottom;
    return AnimatedPadding(
      duration: const Duration(milliseconds: 150),
      padding: EdgeInsets.only(bottom: bottom),
      child: Material(
        color: Colors.transparent,
        child: Container(
          padding: const EdgeInsets.fromLTRB(16, 16, 16, 24),
          decoration: const BoxDecoration(
            color: Colors.white, borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
            boxShadow: [BoxShadow(color: Color(0x22000000), blurRadius: 20, offset: Offset(0, -6))],
          ),
          child: SafeArea(
            top: false,
            child: Column(mainAxisSize: MainAxisSize.min, crossAxisAlignment: CrossAxisAlignment.start, children: [
              Row(children: [
                const Text('Add Expense', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700)),
                const Spacer(), IconButton(icon: const Icon(Icons.close), onPressed: ()=>Navigator.of(context).pop(false)),
              ]),
              const SizedBox(height: 12),
              TextField(controller: _amountCtrl, keyboardType: TextInputType.number, decoration: const InputDecoration(labelText: 'Amount (₹)', border: OutlineInputBorder())),
              const SizedBox(height: 12),
              _Chips(label:'Category', values:_cats, value:_category, onChanged:(v)=>setState(()=>_category=v)),
              const SizedBox(height: 8),
              _Chips(label:'Mood', values:_moods, value:_mood, onChanged:(v)=>setState(()=>_mood=v)),
              const SizedBox(height: 12),
              TextField(controller: _noteCtrl, decoration: const InputDecoration(labelText: 'Note (optional)', border: OutlineInputBorder())),
              if (_error != null) ...[const SizedBox(height: 10), Text(_error!, style: const TextStyle(color: Colors.red))],
              const SizedBox(height: 16),
              SizedBox(width: double.infinity, child: FilledButton(onPressed: _saving?null:_submit, child: _saving?const SizedBox(height:18,width:18,child:CircularProgressIndicator(strokeWidth:2)):const Text('Save'))),
            ]),
          ),
        ),
      ),
    );
  }
}

class _Chips extends StatelessWidget {
  final String label; final List<String> values; final String value; final ValueChanged<String> onChanged;
  const _Chips({required this.label, required this.values, required this.value, required this.onChanged});
  @override
  Widget build(BuildContext context) {
    return Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
      Text(label, style: const TextStyle(fontWeight: FontWeight.w600)),
      const SizedBox(height: 6),
      Wrap(spacing: 8, children: values.map((v){
        final sel = v==value;
        return ChoiceChip(label: Text(v), selected: sel, onSelected: (_)=>onChanged(v));
      }).toList()),
    ]);
  }
}
