import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Ruam Thai Srang Chati | Government Subsidy System",
  description:
    "A once-only government subsidy platform for claims, eligibility checks, real-time status, and project administration.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
