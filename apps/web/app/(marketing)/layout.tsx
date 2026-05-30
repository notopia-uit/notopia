import Footer from '@notopia-uit/ui/components/landing-footer';
import NavBar from '@notopia-uit/ui/components/landing-navigation-bar';

interface MarketingLayoutProps {
  children: React.ReactNode;
}

export default function MarketingLayout({ children }: MarketingLayoutProps) {
  return (
    <div className="flex min-h-screen flex-col">
      <header className="border-b px-6 py-3">
        <div className="mx-auto max-w-6xl">
          <NavBar />
        </div>
      </header>
      <main className="flex-1">{children}</main>
      <Footer />
    </div>
  );
}
