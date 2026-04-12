import Footer from '@notopia-uit/ui/components/landing-footer';
import NavBar from '@notopia-uit/ui/components/landing-navigation-bar';

interface MarketingLayoutProps {
  children: React.ReactNode;
}

export default function MarketingLayout({ children }: MarketingLayoutProps) {
  return (
    <div className="flex flex-col min-h-screen">
      <header className="container border-b ">
        <div className="flex items-center justify-between  px-2 py-2 space-x-2">
          <NavBar />
        </div>{' '}
      </header>
      <main>{children}</main>
      <Footer />
    </div>
  );
}
