import 'package:flutter/material.dart';
import 'package:supabase_flutter/supabase_flutter.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});
  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _email = TextEditingController();
  bool _sending = false;
  String? _msg;

  Future<void> _magicLink() async {
    final email = _email.text.trim();
    if (email.isEmpty) return;
    setState(()=>_sending=true);
    try {
      await Supabase.instance.client.auth.signInWithOtp(email: email, emailRedirectTo: null);
      setState(()=>_msg = "Check your email for login link.");
    } catch (e) {
      setState(()=>_msg = e.toString());
    } finally {
      setState(()=>_sending=false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text("Sign in")),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(children: [
          TextField(controller: _email, decoration: const InputDecoration(labelText: 'Email', border: OutlineInputBorder())),
          const SizedBox(height: 12),
          SizedBox(width: double.infinity, child: FilledButton(onPressed: _sending?null:_magicLink, child: _sending?const CircularProgressIndicator():const Text('Send Magic Link'))),
          if (_msg != null) Padding(padding: const EdgeInsets.only(top: 12), child: Text(_msg!)),
        ]),
      ),
    );
  }
}
