import 'package:flutter_dotenv/flutter_dotenv.dart';
import '../../../../shared/models/order.dart';

class ReceiptService {
  static String format(Order order) {
    final name = dotenv.get('APP_NAME', fallback: 'SALEFORGE').toUpperCase();
    final buf = StringBuffer()
      ..writeln('================================')
      ..writeln('          $name')
      ..writeln('================================')
      ..writeln()
      ..writeln('Order  : ${order.id}')
      ..writeln('Date   : ${_formatDate(order.createdAt)}')
      ..writeln('Payment: ${order.paymentMethod}')
      ..writeln('--------------------------------')
      ..writeln();

    for (final item in order.items) {
      final line = '${item.name}';
      final qty = 'x${item.quantity}';
      final sub = 'Rp${(item.price * item.quantity).toStringAsFixed(0)}';
      buf.writeln(line);
      buf.writeln('  $qty          $sub');
    }

    buf.writeln();
    buf.writeln('--------------------------------');
    buf.writeln('Total     Rp${order.total.toStringAsFixed(0)}');
    buf.writeln('Status    ${order.status == OrderStatus.paid ? "PAID" : "UNPAID"}');
    buf.writeln();
    buf.writeln('================================');
    buf.writeln('    Thank you!');

    return buf.toString();
  }

  static String _formatDate(DateTime dt) {
    return '${dt.day.toString().padLeft(2, '0')}/${dt.month.toString().padLeft(2, '0')}/${dt.year} ${dt.hour.toString().padLeft(2, '0')}:${dt.minute.toString().padLeft(2, '0')}';
  }
}
