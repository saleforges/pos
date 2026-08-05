import 'package:intl/intl.dart';
import '../../../../shared/models/order.dart';
import '../../../../core/config/translations.dart';
import '../../../../core/config/app_config.dart';

class ReceiptService {
  static String format(
    Order order, {
    Translations? t,
    String customer = '',
    String cashier = '',
    String? branchName,
    String? merchantAddress,
    String? merchantPhone,
    String merchantName = '',
    int paperWidth = AppConfig.printerPaperWidth,
  }) {
    final cpl = paperWidth == 80 ? 48 : 32;
    final buf = StringBuffer();
    final locale = _localeString(t?.locale);

    // --- Header ---
    final headerLines = <String>[];
    if (merchantName.isNotEmpty) headerLines.add(merchantName.toUpperCase());
    if ((branchName ?? '').isNotEmpty) headerLines.add(branchName!);
    if ((merchantAddress ?? '').isNotEmpty) headerLines.add(merchantAddress!);
    if ((merchantPhone ?? '').isNotEmpty) headerLines.add(merchantPhone!);

    buf.writeln(_line('=', cpl));
    for (final line in headerLines) {
      for (final w in _wrap(line, cpl)) {
        buf.writeln(_center(w, cpl));
      }
    }
    buf.writeln(_line('=', cpl));

    // --- Transaction Info ---
    final infoLabels = <String>[
      _r(t, 'receipt_order', 'Order No.'),
      _r(t, 'receipt_date', 'Date'),
      if (cashier.isNotEmpty) _r(t, 'receipt_cashier', 'Cashier'),
      if (customer.isNotEmpty) _r(t, 'receipt_customer', 'Customer'),
      _r(t, 'receipt_payment', 'Payment'),
    ];
    final infoPad = infoLabels.fold(0, (max, l) => l.length > max ? l.length : max);
    final infoValues = <String>[
      order.id,
      _formatDate(order.createdAt, locale),
      if (cashier.isNotEmpty) cashier,
      if (customer.isNotEmpty) customer,
      order.paymentMethod,
    ];
    for (var i = 0; i < infoLabels.length; i++) {
      buf.writeln('${infoLabels[i]}${' ' * (infoPad - infoLabels[i].length)} : ${infoValues[i]}');
    }
    buf.writeln(_line('-', cpl));

    // --- Items ---
    for (final item in order.items) {
      for (final w in _wrap(item.name, cpl)) {
        buf.writeln(w);
      }
      final qtyPrice = '${item.quantity} x ${_nf(item.price, locale)}';
      final totalStr = _nf(item.price * item.quantity, locale);
      final gap = cpl - qtyPrice.length - totalStr.length;
      buf.writeln('$qtyPrice${' ' * (gap > 0 ? gap : 1)}$totalStr');
      buf.writeln();
    }

    // --- Summary ---
    final subtotal = order.items.fold<int>(0, (sum, i) => sum + i.price * i.quantity);
    final hasAdjustments = order.discount > 0 || order.tax > 0 || order.serviceCharge > 0 || order.rounding != 0;

    if (hasAdjustments) {
      buf.writeln(_line('-', cpl));
      buf.writeln(_summaryLine(_r(t, 'receipt_subtotal', 'Subtotal'), _rp(subtotal, locale), cpl));
      if (order.discount > 0) {
        buf.writeln(_summaryLine(_r(t, 'receipt_discount', 'Discount'), '-${_rp(order.discount, locale)}', cpl));
      }
      if (order.tax > 0) {
        buf.writeln(_summaryLine(_r(t, 'receipt_tax', 'Tax'), _rp(order.tax, locale), cpl));
      }
      if (order.serviceCharge > 0) {
        buf.writeln(_summaryLine(_r(t, 'receipt_service_charge', 'Service Charge'), _rp(order.serviceCharge, locale), cpl));
      }
      if (order.rounding != 0) {
        final sign = order.rounding > 0 ? '' : '-';
        buf.writeln(_summaryLine(_r(t, 'receipt_rounding', 'Rounding'), '$sign${_rp(order.rounding.abs(), locale)}', cpl));
      }
    }

    buf.writeln(_line('=', cpl));
    buf.writeln(_summaryLine(_r(t, 'receipt_total', 'TOTAL'), _rp(order.total, locale), cpl));
    buf.writeln(_line('=', cpl));

    // --- Footer ---
    buf.writeln(_center(_r(t, 'receipt_thanks', 'Thank you!'), cpl));
    buf.writeln(_line('=', cpl));

    return buf.toString();
  }

  static String _localeString(AppLocale? locale) {
    return locale == AppLocale.en ? 'en' : 'id';
  }

  static String _r(Translations? t, String key, String fallback) =>
      t?.tr(key) ?? fallback;

  static String _rp(int amount, String locale) {
    return NumberFormat.currency(locale: locale, symbol: 'Rp', decimalDigits: 0).format(amount);
  }

  static String _nf(int amount, String locale) {
    return NumberFormat('#,##0', locale).format(amount);
  }

  static List<String> _wrap(String text, int width) {
    if (text.isEmpty) return [''];
    if (text.length <= width) return [text];
    final result = <String>[];
    final words = text.split(' ');
    var line = '';
    for (final word in words) {
      if (word.length > width) {
        if (line.isNotEmpty) {
          result.add(line);
          line = '';
        }
        for (var i = 0; i < word.length; i += width) {
          result.add(word.substring(i, (i + width) > word.length ? word.length : i + width));
        }
        continue;
      }
      if (line.isEmpty) {
        line = word;
      } else if ('$line $word'.length <= width) {
        line = '$line $word';
      } else {
        result.add(line);
        line = word;
      }
    }
    if (line.isNotEmpty) result.add(line);
    return result;
  }

  static String _line(String ch, int w) => ch * w;

  static String _center(String text, int w) {
    if (text.length >= w) return text;
    final pad = (w - text.length) ~/ 2;
    return '${' ' * pad}$text';
  }

  static String _summaryLine(String label, String value, int width) {
    final gap = width - label.length - value.length;
    return '$label${' ' * (gap > 0 ? gap : 1)}$value';
  }

  static String _formatDate(DateTime dt, String locale) {
    if (locale == 'en') {
      return DateFormat('MM/dd/yyyy HH:mm').format(dt);
    }
    return DateFormat('dd/MM/yyyy HH:mm').format(dt);
  }
}
