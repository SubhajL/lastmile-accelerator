export const metadata = {
  title: 'Mode C Zip Uploader',
  description: 'Zip file upload interface',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
