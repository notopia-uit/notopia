import {
  LandingNavbar,
  LandingHeroSection1,
  LandingHeroSection2,
} from '@notopia-uit/ui/components/landing';

export default function Index() {
  return (
    <main className="min-h-screen bg-white">
      <div className="px-4 py-4">
        <LandingNavbar />
      </div>
      <LandingHeroSection1 />
      <LandingHeroSection2 />
    </main>
  );
}
