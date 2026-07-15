import 'package:flutter/material.dart';
import '../../../../shared/models/order.dart';
import 'receipt_service.dart';
import 'bluetooth_printer_service.dart';

class PrintService {
  static Future<void> printReceipt(BuildContext context, Order order) async {
    final bt = BluetoothPrinterService();
    final receipt = ReceiptService.format(order);

    if (bt.isConnected) {
      try {
        await bt.printBytes(BluetoothPrinterService.receiptToBytes(receipt));
        if (context.mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Receipt printed')),
          );
        }
        return;
      } catch (_) {}
    }

    if (!context.mounted) return;
    _showReceiptDialog(context, receipt);
  }

  static void _showReceiptDialog(BuildContext context, String receipt) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Row(
          children: [
            const Icon(Icons.receipt, size: 24),
            const SizedBox(width: 8),
            const Text('Receipt'),
            const Spacer(),
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Close'),
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
