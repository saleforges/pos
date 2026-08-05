package com.saleforge.pos_mobile

import android.bluetooth.BluetoothAdapter
import android.bluetooth.BluetoothDevice
import android.bluetooth.BluetoothSocket
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import io.flutter.embedding.engine.plugins.FlutterPlugin
import io.flutter.plugin.common.EventChannel
import io.flutter.plugin.common.MethodCall
import io.flutter.plugin.common.MethodChannel
import java.io.IOException
import java.util.UUID

class BluetoothPrinterPlugin : FlutterPlugin, MethodChannel.MethodCallHandler, EventChannel.StreamHandler {
    private lateinit var methodChannel: MethodChannel
    private lateinit var eventChannel: EventChannel
    private lateinit var context: Context
    private var adapter: BluetoothAdapter? = null
    private var btSocket: BluetoothSocket? = null
    private var eventSink: EventChannel.EventSink? = null
    private var receiver: BroadcastReceiver? = null

    override fun onAttachedToEngine(binding: FlutterPlugin.FlutterPluginBinding) {
        methodChannel = MethodChannel(binding.binaryMessenger, "bluetooth_printer/methods")
        methodChannel.setMethodCallHandler(this)
        eventChannel = EventChannel(binding.binaryMessenger, "bluetooth_printer/events")
        eventChannel.setStreamHandler(this)
        context = binding.applicationContext
        adapter = BluetoothAdapter.getDefaultAdapter()
    }

    override fun onDetachedFromEngine(binding: FlutterPlugin.FlutterPluginBinding) {
        methodChannel.setMethodCallHandler(null)
        eventChannel.setStreamHandler(null)
        cleanup()
    }

    override fun onListen(arguments: Any?, events: EventChannel.EventSink?) {
        eventSink = events
    }

    override fun onCancel(arguments: Any?) {
        eventSink = null
    }

    override fun onMethodCall(call: MethodCall, result: MethodChannel.Result) {
        when (call.method) {
            "startScan" -> startScan(result)
            "stopScan" -> stopScan(result)
            "connect" -> connect(call.argument<String>("address"), result)
            "writeData" -> writeData(call.argument<ByteArray>("bytes"), result)
            "disconnect" -> disconnect(result)
            "isConnected" -> result.success(btSocket?.isConnected == true)
            else -> result.notImplemented()
        }
    }

    private fun startScan(result: MethodChannel.Result) {
        val bt = adapter ?: run { result.error("no_bt", "Bluetooth not available", null); return }
        stopReceiver()

        val filter = IntentFilter(BluetoothDevice.ACTION_FOUND).apply {
            addAction(BluetoothAdapter.ACTION_DISCOVERY_FINISHED)
        }

        receiver = object : BroadcastReceiver() {
            override fun onReceive(ctx: Context, intent: Intent) {
                when (intent.action) {
                    BluetoothDevice.ACTION_FOUND -> {
                        val device = intent.getParcelableExtra<BluetoothDevice>(BluetoothDevice.EXTRA_DEVICE) ?: return
                        eventSink?.success(mapOf(
                            "name" to (device.name ?: "Unknown"),
                            "address" to (device.address ?: ""),
                            "type" to device.type
                        ))
                    }
                    BluetoothAdapter.ACTION_DISCOVERY_FINISHED -> {
                        eventSink?.success(mapOf("event" to "scan_done"))
                    }
                }
            }
        }

        context.registerReceiver(receiver, filter)

        bt.bondedDevices?.forEach { device ->
            eventSink?.success(mapOf(
                "name" to (device.name ?: "Unknown"),
                "address" to (device.address ?: ""),
                "type" to device.type,
                "bonded" to true
            ))
        }

        bt.startDiscovery()
        result.success(true)
    }

    private fun stopScan(result: MethodChannel.Result) {
        adapter?.cancelDiscovery()
        stopReceiver()
        eventSink?.success(mapOf("event" to "scan_done"))
        result.success(true)
    }

    private fun connect(address: String?, result: MethodChannel.Result) {
        if (address == null) { result.error("no_address", "Address required", null); return }
        val bt = adapter ?: run { result.error("no_bt", "Bluetooth not available", null); return }

        try {
            val device = bt.getRemoteDevice(address)
            bt.cancelDiscovery()
            val sock = device.createRfcommSocketToServiceRecord(UUID.fromString("00001101-0000-1000-8000-00805f9b34fb"))
            sock.connect()
            btSocket = sock
            result.success(true)
        } catch (e: SecurityException) {
            result.error("no_perm", "Permission denied", null)
        } catch (e: IOException) {
            result.error("conn_fail", e.message ?: "Connection failed", null)
        }
    }

    private fun writeData(bytes: ByteArray?, result: MethodChannel.Result) {
        if (bytes == null) { result.error("no_data", "No data", null); return }
        try {
            val sock = btSocket ?: run { result.error("not_conn", "Not connected", null); return }
            sock.outputStream.write(bytes)
            sock.outputStream.flush()
            result.success(true)
        } catch (e: IOException) {
            result.error("write_fail", e.message, null)
        }
    }

    private fun disconnect(result: MethodChannel.Result) {
        try { btSocket?.close() } catch (_: Exception) {}
        btSocket = null
        result.success(true)
    }

    private fun stopReceiver() {
        try { receiver?.let { context.unregisterReceiver(it) } } catch (_: Exception) {}
        receiver = null
    }

    private fun cleanup() {
        stopReceiver()
        try { btSocket?.close() } catch (_: Exception) {}
        btSocket = null
    }
}
