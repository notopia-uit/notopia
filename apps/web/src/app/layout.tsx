import './global.css';

export const metadata = {
  title: 'Notopia',
  description: 'Utopia of Notes',
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
