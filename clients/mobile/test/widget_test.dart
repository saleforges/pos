import 'package:flutter_test/flutter_test.dart';
import 'package:pos_mobile/app.dart';

void main() {
  testWidgets('App should render login screen', (WidgetTester tester) async {
    await tester.pumpWidget(const PosMobileApp());
    expect(find.text('POS Mobile'), findsOneWidget);
    expect(find.text('Sign in to your account'), findsOneWidget);
  });
}
