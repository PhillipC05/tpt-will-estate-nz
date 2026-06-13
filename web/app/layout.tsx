import React from "react";

export const metadata = {
  title: 'Digital Will & Estate',
  description: 'Secure digital wills platform'
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}

