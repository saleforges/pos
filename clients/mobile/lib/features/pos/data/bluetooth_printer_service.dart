import 'dart:async';
import 'dart:convert';
import 'dart:io' show Platform;
import 'package:flutter/services.dart';

class BtPrinter {
  final String name;
  final String address;
  final bool bonded;
  BtPrinter({required this.name, required this.address, this.bonded = false});
}

class BluetoothPrinterService {
  static final BluetoothPrinterService _instance = BluetoothPrinterService._();
  factory BluetoothPrinterService() => _instance;
  BluetoothPrinterService._();

  static const _channel = MethodChannel('bluetooth_printer/methods');
  static const _eventChannel = EventChannel('bluetooth_printer/events');

  static bool get supported => Platform.isAndroid;

  bool _connected = false;
  bool get isConnected => _connected;

  final _devices = <BtPrinter>[];
  List<BtPrinter> get devices => List.unmodifiable(_devices);

  final _onDevicesChanged = StreamController<List<BtPrinter>>.broadcast();
  Stream<List<BtPrinter>> get onDevicesChanged => _onDevicesChanged.stream;

  StreamSubscription? _eventSub;

  void _onEvent(dynamic event) {
    if (event is! Map) return;
    if (event['event'] == 'scan_done') return;
    _devices.add(BtPrinter(
      name: event['name'] ?? 'Unknown',
      address: event['address'] ?? '',
      bonded: event['bonded'] == true,
    ));
    _onDevicesChanged.add(List.from(_devices));
  }

  Future<void> startScan() async {
    if (!supported) throw Exception('Bluetooth not supported on this platform');
    _devices.clear();
    _onDevicesChanged.add(_devices);
    await _eventSub?.cancel();
    _eventSub = _eventChannel.receiveBroadcastStream().listen(_onEvent);
    try {
      await _channel.invokeMethod('startScan');
    } on PlatformException catch (e) {
      if (e.code == 'bt_off') {
        _eventSub?.cancel();
        _eventSub = null;
      }
      rethrow;
    }
  }

  Future<void> stopScan() async {
    if (!supported) return;
    await _channel.invokeMethod('stopScan');
    await _eventSub?.cancel();
  }

  Future<void> connect(BtPrinter printer) async {
    if (!supported) throw Exception('Bluetooth not supported on this platform');
    await _channel.invokeMethod('connect', {'address': printer.address});
    _connected = true;
  }

  Future<void> printBytes(List<int> bytes) async {
    if (!supported) throw Exception('Bluetooth not supported on this platform');
    try {
      final ok = await _channel.invokeMethod('isConnected');
      if (ok != true) {
        _connected = false;
        throw Exception('Printer not connected');
      }
    } catch (_) {}
    if (!_connected) throw Exception('Printer not connected');
    await _channel.invokeMethod('writeData', Uint8List.fromList(bytes));
  }

  Future<void> disconnect() async {
    if (!supported) return;
    await _channel.invokeMethod('disconnect');
    _connected = false;
  }

  void dispose() {
    _eventSub?.cancel();
    _onDevicesChanged.close();
  }

  static List<int> receiptToBytes(String text) {
    final bytes = <int>[0x1B, 0x40];
    final textBytes = latin1.encode(text.replaceAll('\r\n', '\n'));
    for (final b in textBytes) {
      bytes.add(b);
    }
    bytes.add(0x0A);
    bytes.add(0x1D);
    bytes.add(0x56);
    bytes.add(0x00);
    return bytes;
  }
}
