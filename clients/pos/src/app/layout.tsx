import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "POS - SaleForges",
  description: "Point of Sale system",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link
          href="https://fonts.googleapis.com/css2?family=Manrope:wght@600;700;800&family=Inter:wght@400;500;600&display=swap"
          rel="stylesheet"
        />
      </head>
      <body className="font-body text-neutral-600">{children}</body>
    </html>
  );
}
