import 'dart:convert';
import 'package:flutter/material.dart';
import 'dart:async';
import 'package:permission_handler/permission_handler.dart';
import 'package:readsms/readsms.dart';
import 'package:http/http.dart' as http;

void main() {
  runApp(MyApp());
}

class MyApp extends StatefulWidget {
  const MyApp({Key? key}) : super(key: key);

  @override
  State<MyApp> createState() => _MyAppState();
}

class _MyAppState extends State<MyApp> {
  final _plugin = Readsms();
  bool setServer = false;
  String serverAddress = '';

  @override
  void initState() {
    super.initState();
    getPermission().then((value) {
      if (value) {
        _plugin.read();
        _plugin.smsStream.listen((event) {
          //

        });
      }
    });
  }

  Future<bool> getPermission() async {
    if (await Permission.sms.status == PermissionStatus.granted) {
      return true;
    } else {
      if (await Permission.sms.request() == PermissionStatus.granted) {
        return true;
      } else {
        return false;
      }
    }
  }

  @override
  void dispose() {
    super.dispose();
    _plugin.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowMaterialGrid: false,
      debugShowCheckedModeBanner: false,
      home: Scaffold(
        appBar: AppBar(
          title: const Text('SMS Server'),
        ),
        body: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text('SMS Server: $serverAddress'),
              SizedBox(height: 20),
              !setServer
                  ? OutlinedButton(
                      onPressed: () {
                        setState(() {
                          setServer = !setServer;
                        });
                      },
                      child: Text('初始化服务器'))
                  : Container(
                      width: 300,
                      child: Column(
                        children: [
                          TextField(
                            decoration: InputDecoration(hintText: '请输入服务器地址'),
                            onChanged: (value) {
                              serverAddress = value;
                            },
                          ),
                          SizedBox(height: 20),
                          OutlinedButton(onPressed: () {
                            setState(() {
                              serverAddress = serverAddress;
                              setServer = !setServer;
                            });
                          }, child: Text('连接服务器'))
                        ],
                      ),
                    )
            ],
          ),
        ),
      ),
    );
  }
}


void sendData() async {
  // Define the URL
  var url = '';

  // Define the JSON data
  var jsonData = {
    "body": "this is",
    "sender": "qweqweqweqwe",
    "time": "212010212"
  };

  // Encode the JSON data
  var body = jsonEncode(jsonData);

  try {
    // Make the POST request
    var response = await http.post(
      url as Uri,
      headers: {"Content-Type": "application/json"},
      body: body,
    );

    // Check if the request was successful
    if (response.statusCode == 200) {
      print('Success! Response: ${response.body}');
    } else {
      print('Request failed with status: ${response.statusCode}');
    }
  } catch (e) {
    print('Error sending data: $e');
  }
}
