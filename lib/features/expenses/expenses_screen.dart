import 'dart:convert';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/providers.dart';
import '../../core/models/expense.dart';
import 'add_expense_sheet.dart';

class ExpensesScreen extends ConsumerStatefulWidget {
  const ExpensesScreen({super.key});
  @override
  ConsumerState<ExpensesScreen> createState() => _ExpensesScreenState();
}

class _ExpensesScreenState extends ConsumerState<ExpensesScreen> {
  final _dateFmt = DateFormat('yyyy-MM-dd');
  bool _loading = false;
  String? _error;
  List<Expense> _items = [];
  DateTime _from = DateTime.now().subtract(const Duration(days: 7));
  DateTime _to = DateTime.now();

  @override
  void initState() {
    super.initState();
    _fetch();
  }

  Future<void> _fetch() async {
    setState(() { _loading = true; _error = null; });
    final dio = ref.read(apiProvider);
    try {
      final resp = await dio.get('/api/expenses', queryParameters: {
        'from': _dateFmt.format(_from),
        'to': _dateFmt.format(_to),
      });
      final data = resp.data as List? ?? [];
      setState(() {
        _items = data.map((e) => Expense.fromJson(Map<String, dynamic>.from(e))).toList();
      });
    } on DioException catch (e) {
      setState(() { _error = e.response?.data?.toString() ?? e.message; });
    } catch (e) {
      setState(() { _error = e.toString(); });
    } finally {
      setState(() { _loading = false; });
    }
  }

  Future<void> _openAdd() async {
    final created = await showModalBottomSheet<bool>(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      backgroundColor: Colors.transparent,
      builder: (_) => const AddExpenseSheet(),
    );
    if (created == true) { _fetch(); }
  }

  int get _totalCents => _items.fold(0, (sum, e) => sum + e.amountCents);

  @override
  Widget build(BuildContext context) {
    final total = (_totalCents / 100).toStringAsFixed(2);
    return Scaffold(
      appBar: AppBar(
        title: const Text('Expenses'),
        actions: [
          IconButton(onPressed: _fetch, icon: const Icon(Icons.refresh)),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: _openAdd, icon: const Icon(Icons.add), label: const Text('Add'),
      ),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: _loading
            ? const Center(child: CircularProgressIndicator())
            : _error != null
                ? Center(child: Text(_error!))
                : _items.isEmpty
                    ? const Center(child: Text('No expenses yet'))
                    : Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                        _Summary(label: 'Total', value: '₹$total'),
                        const SizedBox(height: 12),
                        Expanded(child: _List(items: _items)),
                      ]),
      ),
    );
  }
}

class _Summary extends StatelessWidget {
  final String label, value;
  const _Summary({required this.label, required this.value});
  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white, borderRadius: BorderRadius.circular(16),
        boxShadow: const [BoxShadow(color: Color(0x11000000), blurRadius: 16, offset: Offset(0, 8))],
      ),
      child: Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children: [
        Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text(label, style: Theme.of(context).textTheme.bodyMedium),
          const SizedBox(height: 6),
          Text(value, style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 22)),
        ]),
        const Icon(Icons.receipt_long_outlined),
      ]),
    );
  }
}

class _List extends StatelessWidget {
  final List<Expense> items;
  const _List({required this.items});
  @override
  Widget build(BuildContext context) {
    return ListView.separated(
      itemCount: items.length,
      separatorBuilder: (_, __) => const SizedBox(height: 10),
      itemBuilder: (_, i) {
        final e = items[i];
        final amount = (e.amountCents / 100).toStringAsFixed(2);
        return Container(
          padding: const EdgeInsets.all(14),
          decoration: BoxDecoration(
            color: Colors.white, borderRadius: BorderRadius.circular(14),
            boxShadow: const [BoxShadow(color: Color(0x0F000000), blurRadius: 12, offset: Offset(0, 6))],
          ),
          child: Row(children: [
            CircleAvatar(radius: 18, backgroundColor: const Color(0xFFF2F4F7),
              child: Text(e.category.isNotEmpty ? e.category[0].toUpperCase() : '?',
                style: const TextStyle(fontWeight: FontWeight.w700))),
            const SizedBox(width: 12),
            Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Text(e.category, style: const TextStyle(fontWeight: FontWeight.w600)),
              if ((e.note ?? '').isNotEmpty) Text(e.note!, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(color: Colors.black54)),
            ])),
            Column(crossAxisAlignment: CrossAxisAlignment.end, children: [
              Text('₹$amount', style: const TextStyle(fontWeight: FontWeight.w600)),
              Text(DateFormat('MMM d').format(e.spentAt), style: const TextStyle(color: Colors.black54)),
            ]),
          ]),
        );
      },
    );
  }
}
