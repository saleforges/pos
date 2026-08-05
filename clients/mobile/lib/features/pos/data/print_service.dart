import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../shared/models/order.dart';
import '../../../../core/config/translations.dart';
import '../../auth/presentation/providers/auth_provider.dart';
import 'receipt_service.dart';
import 'bluetooth_printer_service.dart';

class PrintService {
  static Future<void> printReceipt(
    BuildContext context,
    Order order, {
    String customer = '',
  }) async {
    final container = ProviderScope.containerOf(context, listen: false);
    final t = container.read(translationsProvider);
    final auth = container.read(authProvider);

    final cashier = auth.user?.username ?? '';
    final branch = auth.selectedBranch?.name;
    final merchantName = auth.user?.roles.firstOrNull?.merchant.name ?? '';

    final bt = BluetoothPrinterService();
    final receipt = ReceiptService.format(
      order,
      t: t,
      customer: customer,
      cashier: cashier,
      branchName: branch,
      merchantName: merchantName,
    );

    if (bt.isConnected) {
      try {
        await bt.printBytes(BluetoothPrinterService.receiptToBytes(receipt));
        if (context.mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(t.tr('receipt_printed'))),
          );
        }
        return;
      } catch (e, stack) {
        debugPrint('Print error: $e\n$stack');
        if (context.mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('Cetak gagal: $e'),
              backgroundColor: Colors.red.shade400,
            ),
          );
        }
      }
    }

    if (!context.mounted) return;
    _showReceiptDialog(context, receipt, t);
  }

  static void _showReceiptDialog(
      BuildContext context, String receipt, Translations t) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Row(
          children: [
            const Icon(Icons.receipt, size: 24),
            const SizedBox(width: 8),
            Text(t.tr('receipt')),
            const Spacer(),
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: Text(t.tr('close')),
            ),
          ],
        ),
        content: SingleChildScrollView(
          child: Container(
            width: 280,
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(4),
              border: Border.all(color: Colors.grey.shade200),
            ),
            child: Text(
              receipt,
              style: const TextStyle(
                fontFamily: 'monospace',
                fontSize: 11,
                height: 1.3,
                color: Colors.black87,
              ),
            ),
          ),
        ),
      ),
    );
  }
}
