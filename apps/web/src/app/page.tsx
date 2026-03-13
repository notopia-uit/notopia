import {
  LandingHeroSection1,
  LandingHeroSection2,
  LandingNavbar,
} from '@notopia-uit/ui/components/landing';

export default function Index() {
  return (
    <main className="min-h-screen bg-white">
      <div className="p-4">
        <LandingNavbar />
      </div>
      <LandingHeroSection1 />
      <LandingHeroSection2 />
    </main>
  );
}
