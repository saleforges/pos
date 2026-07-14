import 'package:flutter/material.dart';
import '../../../../shared/models/order.dart';
import 'receipt_service.dart';

class PrintService {
  /// Show receipt preview dialog (fallback when no printer).
  /// When a printer is available, replace this with actual Bluetooth printing.
  static Future<void> printReceipt(BuildContext context, Order order) {
    return showDialog(
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
              ReceiptService.format(order),
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
